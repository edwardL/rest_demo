package lock

import "sync"

// Semaphore 信号量
type Semaphore struct {
	sync.Mutex
	cond    *sync.Cond
	permits int
}

// NewSemaphore 初始化信号量, 设置允许的并发数
func NewSemaphore(permits int) *Semaphore {
	return &Semaphore{
		permits: permits,
		cond:    sync.NewCond(&sync.Mutex{}),
	}
}

// Acquire 获取信号量,如果信号量不足 则阻塞
func (s *Semaphore) Acquire() {
	s.cond.L.Lock()
	for s.permits == 0 {
		s.cond.Wait()
	}
	s.permits--
	s.cond.L.Unlock()
}

func (s *Semaphore) Release() {
	s.cond.L.Lock()
	s.permits++
	s.cond.Signal()
	s.cond.L.Unlock()
}
