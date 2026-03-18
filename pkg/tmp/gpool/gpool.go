package gpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrNilTask 表示提交了空任务。
	ErrNilTask = errors.New("gpool: nil task")
	// ErrClosed 表示协程池已关闭，不再接收新任务。
	ErrClosed = errors.New("gpool: pool closed")
	// ErrQueueFull 表示队列已满且当前策略不允许阻塞等待。
	ErrQueueFull = errors.New("gpool: queue full")
)

// QueuePolicy 表示队列满时的处理策略。
type QueuePolicy int

const (
	// QueuePolicyBlock 在队列满时阻塞等待，直到入队成功或 ctx/池关闭。
	QueuePolicyBlock QueuePolicy = iota
	// QueuePolicyReject 在队列满时立即返回 ErrQueueFull。
	QueuePolicyReject
	// QueuePolicyCallerRuns 在队列满时由调用方直接执行任务。
	QueuePolicyCallerRuns
)

// Options 是协程池配置。
type Options struct {
	MaxWorkers   int
	MinWorkers   int
	QueueSize    int
	IdleTimeout  time.Duration
	QueuePolicy  QueuePolicy
	PanicHandler func(any)
}

// DefaultOptions 返回默认配置。
func DefaultOptions() Options {
	const maxWorkers = 32
	return Options{
		MaxWorkers:  maxWorkers,
		MinWorkers:  1,
		QueueSize:   maxWorkers * 8,
		IdleTimeout: 10 * time.Second,
		QueuePolicy: QueuePolicyBlock,
	}
}

// ZeroQueueOptions 返回零队列预设。
//
// 适合希望尽快背压、尽量不积压任务的场景。
func ZeroQueueOptions() Options {
	const maxWorkers = 32
	return Options{
		MaxWorkers:  maxWorkers,
		MinWorkers:  1,
		QueueSize:   0,
		IdleTimeout: 10 * time.Second,
		QueuePolicy: QueuePolicyBlock,
	}
}

// NonBlockingOptions 返回非阻塞拒绝预设。
//
// 适合更关注调用方时延、无法接受阻塞等待的场景。
func NonBlockingOptions() Options {
	const maxWorkers = 32
	return Options{
		MaxWorkers:  maxWorkers,
		MinWorkers:  1,
		QueueSize:   maxWorkers * 2,
		IdleTimeout: 10 * time.Second,
		QueuePolicy: QueuePolicyReject,
	}
}

// BurstOptions 返回突发流量友好预设。
//
// 适合任务较轻、允许短时堆积以换取更高吞吐的场景。
func BurstOptions() Options {
	const maxWorkers = 32
	return Options{
		MaxWorkers:  maxWorkers,
		MinWorkers:  4,
		QueueSize:   maxWorkers * 32,
		IdleTimeout: 15 * time.Second,
		QueuePolicy: QueuePolicyBlock,
	}
}

// Stats 为协程池运行统计。
type Stats struct {
	Workers   int
	Submitted int64
	Completed int64
	Panics    int64
	Rejected  int64
	Queued    int64
	Running   int64
}

type task struct {
	fn func()
}

// Pool 是一个支持动态伸缩的通用协程池。
type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts Options

	taskCh chan task

	mu        sync.Mutex
	accepting bool
	closedCh  chan struct{}

	workersWG sync.WaitGroup
	workers   int32

	submitted atomic.Int64
	completed atomic.Int64
	panics    atomic.Int64
	rejected  atomic.Int64
	queued    atomic.Int64
	running   atomic.Int64

	stateCh chan struct{}
}

// New 创建一个新的协程池。
func New(ctx context.Context, opts Options) (*Pool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.MaxWorkers <= 0 {
		return nil, fmt.Errorf("gpool: invalid MaxWorkers=%d", opts.MaxWorkers)
	}
	if opts.MinWorkers < 0 {
		return nil, fmt.Errorf("gpool: invalid MinWorkers=%d", opts.MinWorkers)
	}
	if opts.MinWorkers > opts.MaxWorkers {
		return nil, fmt.Errorf("gpool: MinWorkers(%d) > MaxWorkers(%d)", opts.MinWorkers, opts.MaxWorkers)
	}
	if opts.QueueSize < 0 {
		return nil, fmt.Errorf("gpool: invalid QueueSize=%d", opts.QueueSize)
	}
	if opts.QueuePolicy < QueuePolicyBlock || opts.QueuePolicy > QueuePolicyCallerRuns {
		return nil, fmt.Errorf("gpool: invalid QueuePolicy=%d", opts.QueuePolicy)
	}
	if opts.IdleTimeout <= 0 {
		opts.IdleTimeout = 10 * time.Second
	}

	pCtx, cancel := context.WithCancel(ctx)
	p := &Pool{
		ctx:       pCtx,
		cancel:    cancel,
		opts:      opts,
		taskCh:    make(chan task, opts.QueueSize),
		accepting: true,
		closedCh:  make(chan struct{}),
		stateCh:   make(chan struct{}, 1),
	}

	for i := 0; i < opts.MinWorkers; i++ {
		p.startWorker()
	}

	return p, nil
}

// NewDefault 使用默认配置创建协程池。
func NewDefault(ctx context.Context) (*Pool, error) {
	return New(ctx, DefaultOptions())
}

// Submit 提交一个任务。
func (p *Pool) Submit(ctx context.Context, fn func()) error {
	if fn == nil {
		p.rejected.Add(1)
		return ErrNilTask
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if !p.isAccepting() {
		p.rejected.Add(1)
		return ErrClosed
	}

	p.tryScaleUp()

	t := task{fn: fn}
	if p.trySubmitFast(t) {
		return nil
	}

	switch p.opts.QueuePolicy {
	case QueuePolicyReject:
		p.rejected.Add(1)
		return ErrQueueFull
	case QueuePolicyCallerRuns:
		p.submitted.Add(1)
		p.running.Add(1)
		p.notifyStateChange()
		defer func() {
			if r := recover(); r != nil {
				p.panics.Add(1)
				if p.opts.PanicHandler != nil {
					p.opts.PanicHandler(r)
				}
			}
			p.running.Add(-1)
			p.completed.Add(1)
			p.notifyStateChange()
		}()
		fn()
		return nil
	}

	select {
	case <-ctx.Done():
		p.rejected.Add(1)
		return ctx.Err()
	case <-p.closedCh:
		p.rejected.Add(1)
		return ErrClosed
	case <-p.ctx.Done():
		p.rejected.Add(1)
		return ErrClosed
	case p.taskCh <- t:
		p.submitted.Add(1)
		p.queued.Add(1)
		p.notifyStateChange()
		return nil
	}
}

// SubmitTry 尝试立即提交任务，不会因队列已满而阻塞。
//
// 若当前无法立即入队，则返回 ErrQueueFull。
func (p *Pool) SubmitTry(ctx context.Context, fn func()) error {
	if fn == nil {
		p.rejected.Add(1)
		return ErrNilTask
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		p.rejected.Add(1)
		return err
	}

	if !p.isAccepting() {
		p.rejected.Add(1)
		return ErrClosed
	}

	p.tryScaleUp()
	if p.trySubmitFast(task{fn: fn}) {
		return nil
	}

	select {
	case <-p.closedCh:
		p.rejected.Add(1)
		return ErrClosed
	case <-p.ctx.Done():
		p.rejected.Add(1)
		return ErrClosed
	default:
		p.rejected.Add(1)
		return ErrQueueFull
	}
}

// SubmitWait 提交任务并等待任务执行结束（包含 fn 返回）。
//
// 若提交失败，直接返回提交错误。
// 若等待过程中 ctx 超时/取消，返回 ctx.Err()；此时任务可能仍在执行。
func (p *Pool) SubmitWait(ctx context.Context, fn func()) error {
	if fn == nil {
		return ErrNilTask
	}
	if ctx == nil {
		ctx = context.Background()
	}

	done := make(chan struct{})
	err := p.Submit(ctx, func() {
		defer close(done)
		fn()
	})
	if err != nil {
		return err
	}

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close 关闭接收入口，不再接收新任务。
func (p *Pool) Close() {
	p.mu.Lock()
	if p.accepting {
		p.accepting = false
		close(p.closedCh)
	}
	p.mu.Unlock()
}

// Shutdown 优雅关闭协程池。
//
// timeout == 0：立即强制关闭并返回 false。
// timeout < 0：无限等待任务清空后关闭并返回 true。
// timeout > 0：在超时内等待清空，超时则强制关闭并返回 false。
func (p *Pool) Shutdown(timeout time.Duration) bool {
	p.Close()

	if timeout == 0 {
		p.cancel()
		p.workersWG.Wait()
		return false
	}

	if timeout < 0 {
		for {
			if p.drained() {
				p.cancel()
				p.workersWG.Wait()
				return true
			}
			if !p.waitStateChange(nil) {
				p.cancel()
				p.workersWG.Wait()
				return false
			}
		}
	}

	deadline := time.Now().Add(timeout)
	for {
		if p.drained() {
			p.cancel()
			p.workersWG.Wait()
			return true
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			p.cancel()
			p.workersWG.Wait()
			return false
		}
		if !p.waitStateChange(&remaining) {
			if p.drained() {
				p.cancel()
				p.workersWG.Wait()
				return true
			}
		}
	}
}

// Stats 返回当前统计信息快照。
func (p *Pool) Stats() Stats {
	return Stats{
		Workers:   int(atomic.LoadInt32(&p.workers)),
		Submitted: p.submitted.Load(),
		Completed: p.completed.Load(),
		Panics:    p.panics.Load(),
		Rejected:  p.rejected.Load(),
		Queued:    p.queued.Load(),
		Running:   p.running.Load(),
	}
}

// WaitIdle 等待协程池进入空闲状态（queued=0 且 running=0）。
func (p *Pool) WaitIdle(timeout time.Duration) bool {
	if timeout == 0 {
		return p.drained()
	}
	if timeout < 0 {
		for {
			if p.drained() {
				return true
			}
			if !p.waitStateChange(nil) {
				return p.drained()
			}
		}
	}

	deadline := time.Now().Add(timeout)
	for {
		if p.drained() {
			return true
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return p.drained()
		}
		if !p.waitStateChange(&remaining) {
			return p.drained()
		}
	}
}

func (p *Pool) startWorker() {
	atomic.AddInt32(&p.workers, 1)
	p.workersWG.Add(1)
	go p.runWorker()
}

func (p *Pool) runWorker() {
	defer func() {
		atomic.AddInt32(&p.workers, -1)
		p.workersWG.Done()
		p.notifyStateChange()
	}()

	idleTimer := time.NewTimer(p.opts.IdleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-idleTimer.C:
			if int(atomic.LoadInt32(&p.workers)) > p.opts.MinWorkers && p.queued.Load() == 0 {
				return
			}
			idleTimer.Reset(p.opts.IdleTimeout)
		case t := <-p.taskCh:
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			p.queued.Add(-1)
			p.running.Add(1)
			p.notifyStateChange()

			func() {
				defer func() {
					if r := recover(); r != nil {
						p.panics.Add(1)
						if p.opts.PanicHandler != nil {
							p.opts.PanicHandler(r)
						}
					}
					p.running.Add(-1)
					p.completed.Add(1)
					p.notifyStateChange()
				}()
				t.fn()
			}()

			idleTimer.Reset(p.opts.IdleTimeout)
		}
	}
}

func (p *Pool) tryScaleUp() {
	for {
		workers := int(atomic.LoadInt32(&p.workers))
		if workers >= p.opts.MaxWorkers {
			return
		}
		if p.queued.Load() < int64(workers) {
			return
		}

		p.mu.Lock()
		workers = int(atomic.LoadInt32(&p.workers))
		if workers < p.opts.MaxWorkers && p.queued.Load() >= int64(workers) {
			p.startWorker()
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()
		return
	}
}

func (p *Pool) trySubmitFast(t task) bool {
	select {
	case <-p.closedCh:
		p.rejected.Add(1)
		return false
	case <-p.ctx.Done():
		p.rejected.Add(1)
		return false
	case p.taskCh <- t:
		p.submitted.Add(1)
		p.queued.Add(1)
		p.notifyStateChange()
		return true
	default:
		return false
	}
}

func (p *Pool) isAccepting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.accepting
}

func (p *Pool) drained() bool {
	return p.queued.Load() == 0 && p.running.Load() == 0
}

func (p *Pool) notifyStateChange() {
	select {
	case p.stateCh <- struct{}{}:
	default:
	}
}

func (p *Pool) waitStateChange(timeout *time.Duration) bool {
	if timeout == nil {
		select {
		case <-p.stateCh:
			return true
		case <-p.ctx.Done():
			return false
		}
	}

	t := time.NewTimer(*timeout)
	defer t.Stop()
	select {
	case <-p.stateCh:
		return true
	case <-p.ctx.Done():
		return false
	case <-t.C:
		return false
	}
}
