package gpool

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

// shouldRunPerfTests 控制性能测试开关，默认关闭以降低环境抖动影响。
func shouldRunPerfTests() bool {
	return os.Getenv("HHIT_RUN_PERF_TESTS") == "1"
}

// TestPoolLatencyProfile 采集协程池任务完成时延画像。
func TestPoolLatencyProfile(t *testing.T) {
	if testing.Short() || !shouldRunPerfTests() {
		t.Skip("skip performance profile test")
	}

	p, err := New(context.Background(), Options{MaxWorkers: 16, MinWorkers: 2, QueueSize: 4096, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(2 * time.Second)

	const n = 5000
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		err = p.Submit(context.Background(), func() {
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
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting all tasks")
	}

	t.Logf("gpool latency profile: tasks=%d total=%s", n, time.Since(start))
}

// BenchmarkPoolSubmitAndRun 基准测试：串行提交并等待执行完成。
func BenchmarkPoolSubmitAndRun(b *testing.B) {
	p, err := New(context.Background(), Options{MaxWorkers: 8, MinWorkers: 1, QueueSize: 1024, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err != nil {
		b.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(2 * time.Second)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		done := make(chan struct{}, 1)
		if err = p.Submit(context.Background(), func() { done <- struct{}{} }); err != nil {
			b.Fatalf("Submit failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			b.Fatal("timeout waiting task")
		}
	}
}

// BenchmarkPoolConcurrentSubmit 基准测试：并发提交吞吐。
func BenchmarkPoolConcurrentSubmit(b *testing.B) {
	p, err := New(context.Background(), Options{MaxWorkers: 32, MinWorkers: 4, QueueSize: 65536, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
	if err != nil {
		b.Fatalf("New failed: %v", err)
	}
	defer p.Shutdown(2 * time.Second)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{}, 1)
			if err = p.Submit(context.Background(), func() { done <- struct{}{} }); err != nil {
				b.Fatalf("Submit failed: %v", err)
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				b.Fatal("timeout waiting task")
			}
		}
	})
}

// BenchmarkPoolQueuePolicies 基准测试：不同队列策略下的提交成本。
func BenchmarkPoolQueuePolicies(b *testing.B) {
	b.Run("reject", func(b *testing.B) {
		p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyReject})
		if err != nil {
			b.Fatalf("New failed: %v", err)
		}
		defer p.Shutdown(2 * time.Second)

		block := make(chan struct{})
		started := make(chan struct{}, 1)
		if err = p.Submit(context.Background(), func() {
			started <- struct{}{}
			<-block
		}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			b.Fatal("warmup task should start")
		}
		if err = p.Submit(context.Background(), func() {}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = p.Submit(context.Background(), func() {})
			if err != nil && err != ErrQueueFull {
				b.Fatalf("unexpected error: %v", err)
			}
		}
		b.StopTimer()
		close(block)
	})

	b.Run("caller_runs", func(b *testing.B) {
		p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyCallerRuns})
		if err != nil {
			b.Fatalf("New failed: %v", err)
		}
		defer p.Shutdown(2 * time.Second)

		block := make(chan struct{})
		started := make(chan struct{}, 1)
		if err = p.Submit(context.Background(), func() {
			started <- struct{}{}
			<-block
		}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			b.Fatal("warmup task should start")
		}
		if err = p.Submit(context.Background(), func() {}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err = p.Submit(context.Background(), func() {}); err != nil {
				b.Fatalf("Submit failed: %v", err)
			}
		}
		b.StopTimer()
		close(block)
	})

	b.Run("submit_try_full", func(b *testing.B) {
		p, err := New(context.Background(), Options{MaxWorkers: 1, MinWorkers: 1, QueueSize: 1, IdleTimeout: time.Second, QueuePolicy: QueuePolicyBlock})
		if err != nil {
			b.Fatalf("New failed: %v", err)
		}
		defer p.Shutdown(2 * time.Second)

		block := make(chan struct{})
		started := make(chan struct{}, 1)
		if err = p.Submit(context.Background(), func() {
			started <- struct{}{}
			<-block
		}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			b.Fatal("warmup task should start")
		}
		if err = p.Submit(context.Background(), func() {}); err != nil {
			b.Fatalf("warmup submit failed: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = p.SubmitTry(context.Background(), func() {})
			if err != nil && err != ErrQueueFull {
				b.Fatalf("unexpected error: %v", err)
			}
		}
		b.StopTimer()
		close(block)
	})
}
