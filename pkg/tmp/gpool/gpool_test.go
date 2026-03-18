package gpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewInvalidOptions 验证非法配置会返回错误。
func TestNewInvalidOptions(t *testing.T) {
	ctx := context.Background()

	_, err := New(ctx, Options{MaxWorkers: 0, MinWorkers: 0, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err == nil {
		t.Fatal("expect error for MaxWorkers=0")
	}

	_, err = New(ctx, Options{MaxWorkers: 1, MinWorkers: 2, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err == nil {
		t.Fatal("expect error for MinWorkers > MaxWorkers")
	}

	_, err = New(ctx, Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicy(99)})
	if err == nil {
		t.Fatal("expect error for invalid QueuePolicy")
	}
}

// TestOptionPresets 验证预设配置符合预期。
func TestOptionPresets(t *testing.T) {
	def := DefaultOptions()
	if def.QueueSize != def.MaxWorkers*8 {
		t.Fatalf("default queue size mismatch: got %d want %d", def.QueueSize, def.MaxWorkers*8)
	}
	if def.QueuePolicy != QueuePolicyBlock {
		t.Fatalf("default queue policy mismatch: %v", def.QueuePolicy)
	}

	zero := ZeroQueueOptions()
	if zero.QueueSize != 0 {
		t.Fatalf("zero queue preset mismatch: %d", zero.QueueSize)
	}

	nonBlocking := NonBlockingOptions()
	if nonBlocking.QueuePolicy != QueuePolicyReject {
		t.Fatalf("non-blocking queue policy mismatch: %v", nonBlocking.QueuePolicy)
	}

	burst := BurstOptions()
	if burst.QueueSize <= burst.MaxWorkers {
		t.Fatalf("burst queue size should be larger than workers: queue=%d workers=%d", burst.QueueSize, burst.MaxWorkers)
	}
}

// TestPoolSubmitExecutesTasks 验证任务提交后可以被执行。
func TestPoolSubmitExecutesTasks(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 4, MinWorkers: 1, QueueSize: 16, IdleTimeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	const n = 20
	var counter atomic.Int64
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		err = p.Submit(context.Background(), func() {
			counter.Add(1)
			wg.Done()
		})
		if err != nil {
			t.Fatalf("Submit failed at %d: %v", i, err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("tasks not finished in time")
	}

	if counter.Load() != n {
		t.Fatalf("unexpected counter: got %d want %d", counter.Load(), n)
	}

	stats := p.Stats()
	if stats.Submitted != n || stats.Completed != n {
		t.Fatalf("unexpected stats: submitted=%d completed=%d", stats.Submitted, stats.Completed)
	}
}

// TestPoolSubmitNilTask 验证空任务被拒绝。
func TestPoolSubmitNilTask(t *testing.T) {
	p, err := NewDefault(context.Background())
	if err != nil {
		t.Fatalf("NewDefault failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	err = p.Submit(context.Background(), nil)
	if !errors.Is(err, ErrNilTask) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPoolPanicRecovery 验证任务 panic 不会打崩协程池。
func TestPoolPanicRecovery(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 2, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	if err = p.Submit(context.Background(), func() { panic("boom") }); err != nil {
		t.Fatalf("Submit panic task failed: %v", err)
	}

	done := make(chan struct{}, 1)
	if err = p.Submit(context.Background(), func() { done <- struct{}{} }); err != nil {
		t.Fatalf("Submit follow-up task failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pool stopped after panic task")
	}

	if p.Stats().Panics == 0 {
		t.Fatal("panic stats should be greater than 0")
	}
}

// TestPoolCloseRejectsSubmit 验证 Close 后不再接收新任务。
func TestPoolCloseRejectsSubmit(t *testing.T) {
	p, err := NewDefault(context.Background())
	if err != nil {
		t.Fatalf("NewDefault failed: %v", err)
	}

	p.Close()
	err = p.Submit(context.Background(), func() {})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = p.Shutdown(time.Second)
}

// TestPoolShutdownSemantics 验证 Shutdown 的优雅与强制语义。
func TestPoolShutdownSemantics(t *testing.T) {
	t.Run("graceful", func(t *testing.T) {
		p, err := New(context.Background(), Options{MaxWorkers: 2, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		if err = p.Submit(context.Background(), func() { time.Sleep(20 * time.Millisecond) }); err != nil {
			t.Fatalf("Submit failed: %v", err)
		}

		if !p.Shutdown(time.Second) {
			t.Fatal("graceful shutdown should succeed")
		}
	})

	t.Run("forced", func(t *testing.T) {
		p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		start := make(chan struct{}, 1)
		if err = p.Submit(context.Background(), func() {
			start <- struct{}{}
			time.Sleep(120 * time.Millisecond)
		}); err != nil {
			t.Fatalf("Submit failed: %v", err)
		}

		select {
		case <-start:
		case <-time.After(time.Second):
			t.Fatal("task should start")
		}

		if p.Shutdown(0) {
			t.Fatal("Shutdown(0) should return false")
		}
	})
}

// TestPoolWaitIdleSemantics 验证 WaitIdle 的超时语义。
func TestPoolWaitIdleSemantics(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	if !p.WaitIdle(0) {
		t.Fatal("empty pool should be idle")
	}

	if err = p.Submit(context.Background(), func() { time.Sleep(50 * time.Millisecond) }); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if p.WaitIdle(10 * time.Millisecond) {
		t.Fatal("WaitIdle should timeout while running")
	}
	if !p.WaitIdle(time.Second) {
		t.Fatal("WaitIdle should eventually succeed")
	}
}

// TestPoolSubmitWait 验证 SubmitWait 的成功与超时语义。
func TestPoolSubmitWait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p, err := New(context.Background(), Options{MaxWorkers: 2, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		defer p.Shutdown(time.Second)

		if err = p.SubmitWait(context.Background(), func() {
			time.Sleep(10 * time.Millisecond)
		}); err != nil {
			t.Fatalf("SubmitWait should succeed: %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 8, IdleTimeout: 100 * time.Millisecond})
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		defer p.Shutdown(time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()

		err = p.SubmitWait(ctx, func() {
			time.Sleep(100 * time.Millisecond)
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expect context deadline, got: %v", err)
		}
	})
}

// TestPoolQueuePolicyReject 验证拒绝策略在队列满时返回 ErrQueueFull。
func TestPoolQueuePolicyReject(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyReject})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	block := make(chan struct{})
	started := make(chan struct{}, 1)
	if err = p.Submit(context.Background(), func() {
		started <- struct{}{}
		<-block
	}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first task should start")
	}
	if err = p.Submit(context.Background(), func() {}); err != nil {
		t.Fatalf("second submit failed: %v", err)
	}
	if err = p.Submit(context.Background(), func() {}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expect ErrQueueFull, got: %v", err)
	}
	close(block)
	if !p.WaitIdle(time.Second) {
		t.Fatal("pool should become idle")
	}
}

// TestPoolQueuePolicyCallerRuns 验证 caller-runs 策略会在提交方直接执行任务。
func TestPoolQueuePolicyCallerRuns(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyCallerRuns})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	block := make(chan struct{})
	started := make(chan struct{}, 1)
	if err = p.Submit(context.Background(), func() {
		started <- struct{}{}
		<-block
	}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first task should start")
	}
	if err = p.Submit(context.Background(), func() {}); err != nil {
		t.Fatalf("second submit failed: %v", err)
	}

	var ran atomic.Bool
	start := time.Now()
	err = p.Submit(context.Background(), func() {
		ran.Store(true)
	})
	if err != nil {
		t.Fatalf("caller-runs submit failed: %v", err)
	}
	if !ran.Load() {
		t.Fatal("task should run in caller goroutine")
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatal("caller-runs fallback should finish quickly")
	}
	close(block)
}

// TestPoolZeroQueueBlocks 验证零队列预设下会直接形成背压。
func TestPoolZeroQueueBlocks(t *testing.T) {
	p, err := New(context.Background(), ZeroQueueOptions())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	block := make(chan struct{})
	if err = p.Submit(context.Background(), func() { <-block }); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err = p.Submit(ctx, func() {})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expect context deadline under zero-queue backpressure, got: %v", err)
	}
	close(block)
}

// TestPoolSubmitTry 验证 SubmitTry 不阻塞且在不可立即入队时返回 ErrQueueFull。
func TestPoolSubmitTry(t *testing.T) {
	p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	block := make(chan struct{})
	started := make(chan struct{}, 1)
	if err = p.Submit(context.Background(), func() {
		started <- struct{}{}
		<-block
	}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first task should start")
	}
	if err = p.Submit(context.Background(), func() {}); err != nil {
		t.Fatalf("second submit failed: %v", err)
	}

	start := time.Now()
	err = p.SubmitTry(context.Background(), func() {})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expect ErrQueueFull, got: %v", err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("SubmitTry should return quickly without blocking")
	}
	close(block)
}

// TestPoolPanicHandler 验证 PanicHandler 会收到 panic 值。
func TestPoolPanicHandler(t *testing.T) {
	panicCh := make(chan any, 1)
	p, err := New(context.Background(), Options{
		MaxWorkers:   2,
		MinWorkers:   1,
		QueueSize:    8,
		IdleTimeout:  time.Second,
		QueuePolicy:  QueuePolicyBlock,
		PanicHandler: func(v any) { panicCh <- v },
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(time.Second)

	if err = p.Submit(context.Background(), func() { panic("panic-handler-check") }); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	select {
	case v := <-panicCh:
		if v != "panic-handler-check" {
			t.Fatalf("unexpected panic payload: %#v", v)
		}
	case <-time.After(time.Second):
		t.Fatal("panic handler should be called")
	}
}
