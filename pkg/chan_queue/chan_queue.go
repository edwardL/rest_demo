package chan_queue

import (
	"context"
	"log"
	"rest_demo/pkg/sync2"
	"sync/atomic"
)

type ChanQueueOpt[T any] func(r *ChanQueue[T])

// SetChanQueueMaxNum channel 缓冲大小
func SetChanQueueMaxNum[T any](n int) ChanQueueOpt[T] {
	return func(r *ChanQueue[T]) {
		if n > 0 {
			r.chanMaxNum = n
		}
	}
}

// SetChanQueueLimit 并发处理最大数量，0则一个个处理
func SetChanQueueLimit[T any](n int) ChanQueueOpt[T] {
	return func(r *ChanQueue[T]) {
		if n > 1 {
			r.limiter = sync2.NewLimiter(n)
		}
	}
}

// SetChanQueueGracefulClose 优雅关闭时，会阻塞处理剩余队列
func SetChanQueueGracefulClose[T any](gracefulClose bool) ChanQueueOpt[T] {
	return func(r *ChanQueue[T]) {
		r.gracefulClose = gracefulClose
	}
}

// ChanQueue channel 队列，提供并发处理，并发控制，优雅退出功能
type ChanQueue[T any] struct {
	noCopy        noCopy // nolint: structcheck
	processFn     func(T)
	chanMaxNum    int
	limiter       *sync2.Limiter
	gracefulClose bool
	msgChan       chan T
	closeChan     chan struct{}
	overChan      chan struct{}
	state         uint32
}

func NewChanQueue[T any](fn func(T), opts ...ChanQueueOpt[T]) *ChanQueue[T] {
	rq := &ChanQueue[T]{
		chanMaxNum: 1000,
		processFn:  fn,
		closeChan:  make(chan struct{}),
		overChan:   make(chan struct{}),
	}
	for i := range opts {
		opts[i](rq)
	}
	rq.msgChan = make(chan T, rq.chanMaxNum)
	return rq
}

// Push 推入队列，缓冲满了将被阻塞
func (rq *ChanQueue[T]) Push(msg T) error {
	return rq.PushContext(context.Background(), msg)
}

// PushContext 推入队列， ctx可以用来控制队列满了的时候的等待时间
func (rq *ChanQueue[T]) PushContext(ctx context.Context, msg T) error {
	select {
	case <-rq.closeChan:
		return ErrChannelClose
	case <-ctx.Done():
		return ctx.Err()
	default:
		rq.msgChan <- msg
	}
	return nil
}

func (rq *ChanQueue[T]) Start() {
	if !atomic.CompareAndSwapUint32(&rq.state, 0, 1) {
		return
	}
	rq.run()
}

func (rq *ChanQueue[T]) Close() error {
	if atomic.CompareAndSwapUint32(&rq.state, 1, 2) {
		close(rq.closeChan)
		<-rq.overChan
		if rq.limiter != nil {
			rq.limiter.Wait()
		}
	}
	return nil
}

func (rq *ChanQueue[T]) Length() int {
	return len(rq.msgChan)
}

func (rq *ChanQueue[T]) run() {
	for {
		select {
		case <-rq.closeChan:
			l := len(rq.msgChan)
			if rq.gracefulClose {
				log.Printf("[ChanQueue]处理队列剩余数量[%d]", l)
				for i := 0; i < l; i++ {
					rq.worker(<-rq.msgChan)
				}
				log.Printf("[ChanQueue]处理队列剩余数量已全部处理完成")
			}
			rq.overChan <- struct{}{}
			return
		case msg := <-rq.msgChan:
			rq.worker(msg)
		}
	}
}

func (rq *ChanQueue[T]) worker(msg T) {
	if rq.limiter == nil {
		rq.processFn(msg)
	} else {
		rq.limiter.Go(context.Background(), func() {
			rq.processFn(msg)
		})
	}
}
