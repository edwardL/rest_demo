package corn

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// NewCornCtx 创建一个绑定到 ctx 的轻量级内存调度器。
//
// 任务会在单个内部 goroutine 中按到期时间顺序执行。
// 当 ctx 被取消后，调度器将停止接收和执行任务。
func NewCornCtx(ctx context.Context) *CornCtx {
	if ctx == nil {
		ctx = context.Background()
	}
	cCtx, cancel := context.WithCancel(ctx)

	cc := &CornCtx{
		ctx:    cCtx,
		cancel: cancel,
		// 使用缓冲队列，避免回调内再次注册定时任务时，
		// 调度器与上层执行环境之间出现循环等待。
		taskCh: make(chan *cornTaskUnit, 1),
		// stopCh 同样使用缓冲，原因与 taskCh 一致：
		// 回调内可能触发 clearTimeout/clearInterval，
		// 若调度器正等待回调返回，需避免互相阻塞。
		stopCh:  make(chan int64, 1),
		tasks:   make(cornTasks, 0, 32),
		taskID:  make(map[int64]*cornTaskUnit, 32),
		stateCh: make(chan struct{}, 1),
		doneCh:  make(chan struct{}),
	}
	heap.Init(&cc.tasks)
	go cc.async()
	return cc
}

type CornCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
	taskCh chan *cornTaskUnit
	stopCh chan int64
	idSeq  atomic.Int64
	tasks  cornTasks
	taskID map[int64]*cornTaskUnit

	mu      sync.Mutex
	pending int
	stateCh chan struct{}
	doneCh  chan struct{}
}

// TaskHandle 表示一个已提交任务的句柄，可用于后续取消。
type TaskHandle struct {
	cc *CornCtx
	id int64
}

// ID 返回任务句柄对应的任务 id。
func (h TaskHandle) ID() int64 {
	return h.id
}

// Cancel 取消该句柄对应任务。
func (h TaskHandle) Cancel() bool {
	if h.cc == nil {
		return false
	}
	return h.cc.Cancel(h.id)
}

// ScheduleAt 将 fn 调度到绝对时间 t 执行。
// 当调度器已关闭或 fn 为 nil 时返回 false。
func (cc *CornCtx) ScheduleAt(t time.Time, fn func()) bool {
	return cc.pushT(t, fn)
}

// ScheduleAtID 将 fn 调度到绝对时间 t 执行，并返回任务 id。
func (cc *CornCtx) ScheduleAtID(t time.Time, fn func()) (int64, bool) {
	return cc.pushTID(t, fn)
}

// ScheduleAtHandle 将 fn 调度到绝对时间 t 执行，并返回任务句柄。
func (cc *CornCtx) ScheduleAtHandle(t time.Time, fn func()) (TaskHandle, bool) {
	id, ok := cc.pushTID(t, fn)
	if !ok {
		return TaskHandle{}, false
	}
	return TaskHandle{cc: cc, id: id}, true
}

// ScheduleAfter 将 fn 调度为在 td 之后执行。
// 当调度器已关闭或 fn 为 nil 时返回 false。
func (cc *CornCtx) ScheduleAfter(td time.Duration, fn func()) bool {
	return cc.pushTD(td, fn)
}

// ScheduleAfterID 将 fn 调度为在 td 之后执行，并返回任务 id。
func (cc *CornCtx) ScheduleAfterID(td time.Duration, fn func()) (int64, bool) {
	return cc.pushTDID(td, fn)
}

// ScheduleAfterHandle 将 fn 调度为在 td 之后执行，并返回任务句柄。
func (cc *CornCtx) ScheduleAfterHandle(td time.Duration, fn func()) (TaskHandle, bool) {
	id, ok := cc.pushTDID(td, fn)
	if !ok {
		return TaskHandle{}, false
	}
	return TaskHandle{cc: cc, id: id}, true
}

// Cancel 按任务 id 取消一个已调度任务。
func (cc *CornCtx) Cancel(id int64) bool {
	return cc.stop(id)
}

// Close 立即关闭调度器，不再接收新任务。
func (cc *CornCtx) Close() {
	cc.cancel()
}

// Shutdown 优雅关闭调度器。
//
// 会先等待已接收任务完成；若超时未完成，则强制关闭并返回 false。
// timeout == 0 表示立即强制关闭。
// timeout < 0 表示无限等待已接收任务完成后再关闭。
func (cc *CornCtx) Shutdown(timeout time.Duration) bool {
	if timeout == 0 {
		cc.cancel()
		<-cc.doneCh
		return false
	}

	if timeout < 0 {
		_ = cc.WaitAll(-1)
		cc.cancel()
		<-cc.doneCh
		return true
	}

	if !cc.WaitAll(timeout) {
		cc.cancel()
		<-cc.doneCh
		return false
	}

	cc.cancel()
	return cc.WaitDone(timeout)
}

// pushT 将 fn 调度到绝对时间 t 执行。
// 当调度器上下文已取消或 fn 为 nil 时返回 false。
func (cc *CornCtx) pushT(t time.Time, fn func()) bool {
	_, ok := cc.pushTID(t, fn)
	return ok
}

// pushTID 将 fn 调度到绝对时间 t 执行，并返回任务 id。
func (cc *CornCtx) pushTID(t time.Time, fn func()) (int64, bool) {
	if fn == nil {
		return 0, false
	}
	if cc.ctx.Err() != nil {
		return 0, false
	}
	id := cc.idSeq.Add(1)

	select {
	case <-cc.ctx.Done():
		return 0, false
	case cc.taskCh <- &cornTaskUnit{
		id: id,
		t:  t,
		fn: fn,
	}:
		cc.changePending(1)
		return id, true
	}
}

// pushTD 将 fn 调度为在 td 之后执行。
// 当调度器上下文已取消或 fn 为 nil 时返回 false。
func (cc *CornCtx) pushTD(td time.Duration, fn func()) bool {
	_, ok := cc.pushTDID(td, fn)
	return ok
}

// pushTDID 将 fn 调度为在 td 之后执行，并返回任务 id。
func (cc *CornCtx) pushTDID(td time.Duration, fn func()) (int64, bool) {
	if fn == nil {
		return 0, false
	}
	if cc.ctx.Err() != nil {
		return 0, false
	}
	id := cc.idSeq.Add(1)

	select {
	case <-cc.ctx.Done():
		return 0, false
	case cc.taskCh <- &cornTaskUnit{
		id: id,
		t:  time.Now().Add(td),
		fn: fn,
	}:
		cc.changePending(1)
		return id, true
	}
}

// stop 按任务 id 取消一个已调度任务。
// 仅当取消请求成功进入调度器队列时返回 true。
func (cc *CornCtx) stop(id int64) bool {
	if id <= 0 {
		return false
	}
	if cc.ctx.Err() != nil {
		return false
	}

	select {
	case <-cc.ctx.Done():
		return false
	case cc.stopCh <- id:
		return true
	}
}

// WaitAll 等待直到所有已接收任务执行完成。
// timeout == 0 表示只做一次非阻塞状态检查。
// timeout < 0 表示无限等待。
func (cc *CornCtx) WaitAll(timeout time.Duration) bool {
	if timeout == 0 {
		return cc.pendingCount() == 0
	}
	if timeout < 0 {
		for {
			if cc.pendingCount() == 0 {
				return true
			}
			if !cc.waitStateChange(nil) {
				return cc.pendingCount() == 0
			}
		}
	}

	deadline := time.Now().Add(timeout)
	for {
		if cc.pendingCount() == 0 {
			return true
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return cc.pendingCount() == 0
		}
		if !cc.waitStateChange(&remaining) {
			return cc.pendingCount() == 0
		}
	}
}

// WaitIdle 等待无待执行任务，并在 idle 窗口内保持空闲状态。
// timeout == 0 表示只做一次非阻塞状态检查。
// timeout < 0 表示无限等待。
func (cc *CornCtx) WaitIdle(idle, timeout time.Duration) bool {
	if idle <= 0 {
		return cc.WaitAll(timeout)
	}
	if timeout == 0 {
		// 非阻塞检查：仅当当前无待执行任务时返回 true，
		// 不会等待 idle 窗口。
		return cc.pendingCount() == 0
	}

	checkIdleWindow := func() bool {
		if cc.pendingCount() != 0 {
			return false
		}
		t := time.NewTimer(idle)
		defer t.Stop()
		select {
		case <-t.C:
			return cc.pendingCount() == 0
		case <-cc.stateCh:
			return false
		case <-cc.doneCh:
			return true
		}
	}

	if timeout < 0 {
		for {
			if checkIdleWindow() {
				return true
			}
			if !cc.waitStateChange(nil) {
				return checkIdleWindow()
			}
		}
	}

	deadline := time.Now().Add(timeout)
	for {
		if checkIdleWindow() {
			return true
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		if !cc.waitStateChange(&remaining) {
			return checkIdleWindow()
		}
	}
}

// WaitDone 等待调度器 goroutine 完全退出。
// timeout == 0 表示只做一次非阻塞检查。
// timeout < 0 表示无限等待。
func (cc *CornCtx) WaitDone(timeout time.Duration) bool {
	if timeout == 0 {
		select {
		case <-cc.doneCh:
			return true
		default:
			return false
		}
	}
	if timeout < 0 {
		<-cc.doneCh
		return true
	}

	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-cc.doneCh:
		return true
	case <-t.C:
		return false
	}
}

// WaitDoneAndCancel 先等待退出；若超时则先取消调度器，
// 再阻塞等待直到完全退出。
func (cc *CornCtx) WaitDoneAndCancel(timeout time.Duration) bool {
	if cc.WaitDone(timeout) {
		return true
	}
	cc.cancel()
	<-cc.doneCh
	return false
}

func (cc *CornCtx) async() {
	defer close(cc.doneCh)

	tr := time.NewTimer(time.Hour)
	defer tr.Stop()
	if !tr.Stop() {
		select {
		case <-tr.C:
		default:
		}
	}

	var timerCh <-chan time.Time
	var task *cornTaskUnit
	stopped := make(map[int64]struct{})
	closing := false
	for {
		if closing && cc.pendingCount() == 0 {
			return
		}

		task = cc.peek()
		if task == nil {
			timerCh = nil
		} else {
			d := time.Until(task.t)
			if d < 0 {
				d = 0
			}
			resetTimer(tr, d)
			timerCh = tr.C
		}

		select {
		case <-timerCh:
			for {
				task = cc.peek()
				if task == nil || task.t.After(time.Now()) {
					break
				}
				task = cc.pop()
				if task != nil {
					delete(cc.taskID, task.id)
				}
				if _, ok := stopped[task.id]; ok {
					delete(stopped, task.id)
					cc.changePending(-1)
					continue
				}
				if task != nil && task.fn != nil {
					runTask(task.fn)
				}
				cc.changePending(-1)
			}
		case task = <-cc.taskCh:
			if _, ok := stopped[task.id]; ok {
				delete(stopped, task.id)
				cc.changePending(-1)
				continue
			}
			heap.Push(&cc.tasks, task)
			cc.taskID[task.id] = task
		case id := <-cc.stopCh:
			if task, ok := cc.taskID[id]; ok && task.index >= 0 {
				heap.Remove(&cc.tasks, task.index)
				delete(cc.taskID, id)
				cc.changePending(-1)
				continue
			}
			stopped[id] = struct{}{}
		case <-cc.ctx.Done():
			closing = true
		}
	}
}

func (cc *CornCtx) changePending(delta int) {
	cc.mu.Lock()
	cc.pending += delta
	if cc.pending < 0 {
		cc.pending = 0
	}
	cc.mu.Unlock()
	cc.notifyStateChange()
}

func (cc *CornCtx) pendingCount() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.pending
}

func (cc *CornCtx) notifyStateChange() {
	select {
	case cc.stateCh <- struct{}{}:
	default:
	}
}

func (cc *CornCtx) waitStateChange(timeout *time.Duration) bool {
	if timeout == nil {
		select {
		case <-cc.stateCh:
			return true
		case <-cc.doneCh:
			return false
		}
	}
	t := time.NewTimer(*timeout)
	defer t.Stop()
	select {
	case <-cc.stateCh:
		return true
	case <-cc.doneCh:
		return false
	case <-t.C:
		return false
	}
}

func (cc *CornCtx) peek() *cornTaskUnit {
	if len(cc.tasks) == 0 {
		return nil
	}
	return cc.tasks[0]
}

func (cc *CornCtx) pop() *cornTaskUnit {
	if len(cc.tasks) == 0 {
		return nil
	}
	return heap.Pop(&cc.tasks).(*cornTaskUnit)
}

func resetTimer(tr *time.Timer, d time.Duration) {
	if !tr.Stop() {
		select {
		case <-tr.C:
		default:
		}
	}
	tr.Reset(d)
}

func runTask(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// 至少记录 panic 信息
			log.Printf("Task panic recovered: %v", r)
		}
	}()
	fn()
}

type cornTaskUnit struct {
	id    int64
	t     time.Time
	fn    func()
	index int
}

type cornTasks []*cornTaskUnit

// Len 返回当前任务堆的元素数量。
func (c *cornTasks) Len() int {
	return len(*c)
}

// Less 比较两个任务的触发时间先后。
func (c *cornTasks) Less(i, j int) bool {
	return (*c)[i].t.Before((*c)[j].t)
}

// Swap 交换两个任务并同步更新索引。
func (c *cornTasks) Swap(i, j int) {
	(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
	(*c)[i].index = i
	(*c)[j].index = j
}

// Push 向任务堆尾部压入一个任务元素。
func (c *cornTasks) Push(x any) {
	task := x.(*cornTaskUnit)
	task.index = len(*c)
	*c = append(*c, task)
}

// Pop 从任务堆尾部弹出一个任务元素。
func (c *cornTasks) Pop() (v any) {
	old := *c
	n := len(old)
	v = old[n-1]
	if task, ok := v.(*cornTaskUnit); ok {
		task.index = -1
	}
	next := old[:n-1]

	// 在突发流量后释放过大的底层数组，
	// 避免低流量阶段长期持有大内存。
	if cap(next) > 4096 && len(next)*4 < cap(next) {
		shrink := make(cornTasks, len(next))
		copy(shrink, next)
		*c = shrink
		return
	}

	*c = next
	return
}
