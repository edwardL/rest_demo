package corn

import (
	"context"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// shouldRunPerfTests 控制性能测试开关。
// 默认关闭，避免在共享 CI 或高负载机器上出现不稳定结果。
func shouldRunPerfTests() bool {
	return os.Getenv("HHIT_RUN_PERF_TESTS") == "1"
}

// TestCornCtxTimingPrecisionProfile 采集调度精度画像（均值/分位/最大延迟）。
// 该用例默认跳过，仅在显式开启性能测试时运行。
func TestCornCtxTimingPrecisionProfile(t *testing.T) {
	if testing.Short() || !shouldRunPerfTests() {
		t.Skip("skip performance profile test")
	}

	const (
		taskCount = 200
		delay     = 15 * time.Millisecond
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	late := make([]time.Duration, taskCount)
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		i := i
		expected := start.Add(delay + time.Duration(i)*time.Millisecond/5)
		ok := cc.pushT(expected, func() {
			now := time.Now()
			lateness := now.Sub(expected)
			if lateness < 0 {
				lateness = 0
			}
			late[i] = lateness
			wg.Done()
		})
		if !ok {
			t.Fatalf("schedule task %d failed", i)
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
		t.Fatal("timed out waiting scheduled tasks")
	}

	sort.Slice(late, func(i, j int) bool { return late[i] < late[j] })

	var total time.Duration
	for _, v := range late {
		total += v
	}
	mean := total / taskCount
	p50 := late[taskCount*50/100]
	p95 := late[taskCount*95/100]
	max := late[taskCount-1]

	// Report the precision level for current environment.
	t.Logf("cron precision profile: mean=%s p50=%s p95=%s max=%s", mean, p50, p95, max)
	if max < 0 {
		t.Fatalf("invalid latency stat: max=%s", max)
	}
}

// BenchmarkCornCtxScheduleAndRun 基准测试：调度+执行单任务端到端开销。
func BenchmarkCornCtxScheduleAndRun(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		done := make(chan struct{}, 1)
		if !cc.pushTD(0, func() {
			done <- struct{}{}
		}) {
			b.Fatal("pushTD failed")
		}
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			b.Fatal("timeout waiting task")
		}
	}
}

// BenchmarkCornCtxBurstThroughput 基准测试：突发批量任务吞吐表现。
func BenchmarkCornCtxBurstThroughput(b *testing.B) {
	const burst = 1000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(burst)
		base := time.Now().Add(2 * time.Millisecond)

		for j := 0; j < burst; j++ {
			if !cc.pushT(base, func() { wg.Done() }) {
				b.Fatalf("pushT failed at %d", j)
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
			b.Fatalf("burst timeout, burst=%d", burst)
		}
	}

	b.StopTimer()
}

// TestCornCtxConcurrentSubmitPrecision 采集并发生产者场景下的调度精度分布。
// 该用例默认跳过，仅在显式开启性能测试时运行。
func TestCornCtxConcurrentSubmitPrecision(t *testing.T) {
	if testing.Short() || !shouldRunPerfTests() {
		t.Skip("skip performance profile test")
	}

	const (
		producers      = 24
		tasksPerWorker = 80
		totalTasks     = producers * tasksPerWorker
		delay          = 25 * time.Millisecond
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	late := make([]time.Duration, totalTasks)
	base := time.Now()

	var seq atomic.Int64
	var doneWG sync.WaitGroup
	doneWG.Add(totalTasks)

	var submitWG sync.WaitGroup
	submitWG.Add(producers)
	for p := 0; p < producers; p++ {
		go func(worker int) {
			defer submitWG.Done()
			for i := 0; i < tasksPerWorker; i++ {
				idx := int(seq.Add(1) - 1)
				expected := base.Add(delay + time.Duration(worker%8)*time.Millisecond)
				ok := cc.pushT(expected, func() {
					now := time.Now()
					lateness := now.Sub(expected)
					if lateness < 0 {
						lateness = 0
					}
					late[idx] = lateness
					doneWG.Done()
				})
				if !ok {
					t.Errorf("schedule failed worker=%d item=%d", worker, i)
					doneWG.Done()
					return
				}
			}
		}(p)
	}
	submitWG.Wait()

	done := make(chan struct{})
	go func() {
		doneWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("timed out waiting concurrent scheduled tasks")
	}

	sort.Slice(late, func(i, j int) bool { return late[i] < late[j] })

	var total time.Duration
	for _, v := range late {
		total += v
	}
	mean := total / totalTasks
	p50 := late[totalTasks*50/100]
	p95 := late[totalTasks*95/100]
	max := late[totalTasks-1]

	t.Logf("cron concurrent precision: mean=%s p50=%s p95=%s max=%s producers=%d tasks=%d", mean, p50, p95, max, producers, totalTasks)
	if max < 0 {
		t.Fatalf("invalid latency stat: max=%s", max)
	}
}

// TestCornCtxConcurrentTailLatencyLongRun 长时间并发提交下采集尾延迟分布。
// 该用例默认跳过，仅在显式开启性能测试时运行。
func TestCornCtxConcurrentTailLatencyLongRun(t *testing.T) {
	if testing.Short() || !shouldRunPerfTests() {
		t.Skip("skip performance profile test")
	}

	const (
		producers      = 16
		tasksPerWorker = 450
		totalTasks     = producers * tasksPerWorker
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	late := make([]time.Duration, totalTasks)

	var submitWG sync.WaitGroup
	submitWG.Add(producers)

	var doneWG sync.WaitGroup
	doneWG.Add(totalTasks)

	var seq atomic.Int64
	rng := rand.New(rand.NewSource(42))
	var rngMu sync.Mutex

	for p := 0; p < producers; p++ {
		go func() {
			defer submitWG.Done()
			for i := 0; i < tasksPerWorker; i++ {
				idx := int(seq.Add(1) - 1)

				rngMu.Lock()
				jitter := time.Duration(rng.Intn(12)) * time.Millisecond
				rngMu.Unlock()

				expected := time.Now().Add(8*time.Millisecond + jitter)
				if !cc.pushT(expected, func() {
					lateness := time.Since(expected)
					if lateness < 0 {
						lateness = 0
					}
					late[idx] = lateness
					doneWG.Done()
				}) {
					t.Errorf("schedule failed")
					doneWG.Done()
					return
				}

				if i%12 == 0 {
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}

	submitWG.Wait()

	done := make(chan struct{})
	go func() {
		doneWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(12 * time.Second):
		t.Fatal("timed out waiting long-run concurrent tasks")
	}

	stats := calcLatencyStats(late)
	t.Logf("cron long-run tail latency: mean=%s p95=%s p99=%s p99.9=%s max=%s n=%d",
		stats.mean, stats.p95, stats.p99, stats.p999, stats.max, len(late))
	if stats.max < 0 {
		t.Fatalf("invalid latency stat: max=%s", stats.max)
	}
}

// BenchmarkCornCtxConcurrentEnqueue 基准测试：多 goroutine 并发入队成本。
func BenchmarkCornCtxConcurrentEnqueue(b *testing.B) {
	const (
		workers        = 16
		tasksPerWorker = 64
		totalTasks     = workers * tasksPerWorker
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var doneWG sync.WaitGroup
		doneWG.Add(totalTasks)

		base := time.Now().Add(2 * time.Millisecond)
		var submitWG sync.WaitGroup
		submitWG.Add(workers)

		for w := 0; w < workers; w++ {
			go func() {
				defer submitWG.Done()
				for k := 0; k < tasksPerWorker; k++ {
					if !cc.pushT(base, func() { doneWG.Done() }) {
						b.Error("pushT failed")
						return
					}
				}
			}()
		}
		submitWG.Wait()

		done := make(chan struct{})
		go func() {
			doneWG.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			b.Fatalf("concurrent enqueue timeout, tasks=%d", totalTasks)
		}
	}
}

// BenchmarkCornCtxExportedAPI 基准测试：导出 API 的调度与取消开销。
func BenchmarkCornCtxExportedAPI(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	b.ReportAllocs()

	b.Run("schedule_after", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			done := make(chan struct{}, 1)
			if !cc.ScheduleAfter(0, func() { done <- struct{}{} }) {
				b.Fatal("ScheduleAfter failed")
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				b.Fatal("timeout waiting scheduled task")
			}
		}
	})

	b.Run("schedule_handle_cancel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			h, ok := cc.ScheduleAfterHandle(time.Hour, func() {})
			if !ok {
				b.Fatal("ScheduleAfterHandle failed")
			}
			if !h.Cancel() {
				b.Fatal("handle cancel failed")
			}
		}
		if !cc.WaitAll(5 * time.Second) {
			b.Fatal("scheduler did not drain pending tasks")
		}
	})
}

type latencyStats struct {
	mean time.Duration
	p95  time.Duration
	p99  time.Duration
	p999 time.Duration
	max  time.Duration
}

// calcLatencyStats 计算延迟分布统计值。
func calcLatencyStats(values []time.Duration) latencyStats {
	if len(values) == 0 {
		return latencyStats{}
	}

	sorted := make([]time.Duration, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, v := range sorted {
		total += v
	}

	return latencyStats{
		mean: total / time.Duration(len(sorted)),
		p95:  percentile(sorted, 0.95),
		p99:  percentile(sorted, 0.99),
		p999: percentile(sorted, 0.999),
		max:  sorted[len(sorted)-1],
	}
}

// percentile 返回已排序样本中的分位值。
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}
