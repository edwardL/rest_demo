package corn

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCornCtxPushTDExecutesTask 验证延时任务可被正常执行。
func TestCornCtxPushTDExecutesTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	done := make(chan struct{}, 1)

	ok := cc.pushTD(20*time.Millisecond, func() {
		done <- struct{}{}
	})
	if !ok {
		t.Fatal("pushTD should succeed")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scheduled task did not execute")
	}
}

// TestCornCtxExecutesByDueTimeOrder 验证任务按到期时间顺序执行。
func TestCornCtxExecutesByDueTimeOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	var mu sync.Mutex
	order := make([]int, 0, 3)
	done := make(chan struct{}, 3)

	base := time.Now()
	if !cc.pushT(base.Add(80*time.Millisecond), func() {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		done <- struct{}{}
	}) {
		t.Fatal("pushT #1 should succeed")
	}
	if !cc.pushT(base.Add(20*time.Millisecond), func() {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		done <- struct{}{}
	}) {
		t.Fatal("pushT #2 should succeed")
	}
	if !cc.pushT(base.Add(50*time.Millisecond), func() {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		done <- struct{}{}
	}) {
		t.Fatal("pushT #3 should succeed")
	}

	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(700 * time.Millisecond):
			t.Fatalf("timed out waiting for task %d", i+1)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 3 {
		t.Fatalf("unexpected order size: %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("unexpected execution order: %v", order)
	}
}

// TestCornCtxRejectsAfterCancel 验证取消后不再接收新任务。
func TestCornCtxRejectsAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cc := NewCornCtx(ctx)
	cancel()

	if cc.pushTD(10*time.Millisecond, func() {}) {
		t.Fatal("pushTD should fail after cancel")
	}
	if cc.pushT(time.Now().Add(10*time.Millisecond), func() {}) {
		t.Fatal("pushT should fail after cancel")
	}
}

// TestCornCtxRejectsNilTask 验证空函数任务会被拒绝。
func TestCornCtxRejectsNilTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	if cc.pushTD(10*time.Millisecond, nil) {
		t.Fatal("pushTD should reject nil function")
	}
	if cc.pushT(time.Now().Add(10*time.Millisecond), nil) {
		t.Fatal("pushT should reject nil function")
	}
}

// TestCornCtxTaskPanicDoesNotBreakScheduler 验证单任务 panic 不影响调度器后续运行。
func TestCornCtxTaskPanicDoesNotBreakScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	done := make(chan struct{}, 1)

	if !cc.pushTD(10*time.Millisecond, func() {
		panic("boom")
	}) {
		t.Fatal("panic task schedule failed")
	}
	if !cc.pushTD(20*time.Millisecond, func() {
		done <- struct{}{}
	}) {
		t.Fatal("follow-up task schedule failed")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scheduler stopped after panic task")
	}
}

// TestCornCtxStopWaitsForAcceptance 验证 stop 生效且被取消任务不会执行。
func TestCornCtxStopWaitsForAcceptance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	var executed atomic.Bool

	id, ok := cc.pushTDID(60*time.Millisecond, func() {
		executed.Store(true)
	})
	if !ok {
		t.Fatal("pushTDID failed")
	}

	if !cc.stop(id) {
		t.Fatal("stop should be accepted")
	}

	if !cc.WaitIdle(15*time.Millisecond, 300*time.Millisecond) {
		t.Fatal("scheduler should become idle after stop")
	}
	if executed.Load() {
		t.Fatal("stopped task should not run")
	}
}

// TestCornCtxWaitAllTimeoutSemantics 验证 WaitAll 的超时与非阻塞语义。
func TestCornCtxWaitAllTimeoutSemantics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)

	if cc.WaitAll(0) != true {
		t.Fatal("WaitAll(0) should pass when there is no pending task")
	}

	ok := cc.pushTD(40*time.Millisecond, func() {})
	if !ok {
		t.Fatal("pushTD failed")
	}

	if cc.WaitAll(0) {
		t.Fatal("WaitAll(0) should fail while task is pending")
	}
	if cc.WaitAll(10 * time.Millisecond) {
		t.Fatal("WaitAll should timeout before task completion")
	}
	if !cc.WaitAll(200 * time.Millisecond) {
		t.Fatal("WaitAll should succeed after task completion")
	}
}

// TestCornCtxWaitIdleSemantics 验证 WaitIdle 能识别链式新增任务后的空闲状态。
func TestCornCtxWaitIdleSemantics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)

	if !cc.WaitIdle(5*time.Millisecond, 50*time.Millisecond) {
		t.Fatal("empty scheduler should become idle")
	}

	ok := cc.pushTD(10*time.Millisecond, func() {
		_ = cc.pushTD(1*time.Millisecond, func() {})
	})
	if !ok {
		t.Fatal("pushTD failed")
	}

	if !cc.WaitIdle(5*time.Millisecond, 300*time.Millisecond) {
		t.Fatal("WaitIdle should eventually succeed")
	}
}

// TestCornCtxWaitIdleZeroTimeoutIsNonBlocking 验证 WaitIdle 零超时为非阻塞检查。
func TestCornCtxWaitIdleZeroTimeoutIsNonBlocking(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	if !cc.WaitIdle(20*time.Millisecond, 0) {
		t.Fatal("WaitIdle(0) should pass immediately when empty")
	}

	if !cc.pushTD(40*time.Millisecond, func() {}) {
		t.Fatal("pushTD failed")
	}
	if cc.WaitIdle(20*time.Millisecond, 0) {
		t.Fatal("WaitIdle(0) must be non-blocking and fail when pending exists")
	}
}

// TestCornCtxWaitDoneAndCancel 验证 WaitDoneAndCancel 的超时取消路径。
func TestCornCtxWaitDoneAndCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)

	_ = cc.pushTD(200*time.Millisecond, func() {
		time.Sleep(100 * time.Millisecond)
	})

	if cc.WaitDone(0) {
		t.Fatal("WaitDone(0) should be false while running")
	}

	if cc.WaitDoneAndCancel(10 * time.Millisecond) {
		t.Fatal("WaitDoneAndCancel should report timeout path")
	}

	if !cc.WaitDone(0) {
		t.Fatal("scheduler should be done after cancel path")
	}
}

// TestCornCtxPushTDIDReturnsUniqueID 验证 pushTDID 返回的任务 id 单调递增且唯一。
func TestCornCtxPushTDIDReturnsUniqueID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	const n = 6

	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		id, ok := cc.pushTDID(5*time.Millisecond, func() {})
		if !ok {
			t.Fatalf("pushTDID failed at %d", i)
		}
		ids = append(ids, id)
	}

	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("ids should be strictly increasing, got %v", ids)
		}
	}

	if !cc.WaitAll(500 * time.Millisecond) {
		t.Fatal("scheduled tasks should complete")
	}
}

// TestCornCtxStopInvalidID 验证 stop 对非法 id 会直接返回 false。
func TestCornCtxStopInvalidID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	if cc.stop(0) {
		t.Fatal("stop(0) should fail")
	}
	if cc.stop(-1) {
		t.Fatal("stop(-1) should fail")
	}
}

// TestCornCtxExportedScheduleAPI 验证导出调度 API 行为与内部实现一致。
func TestCornCtxExportedScheduleAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	done := make(chan struct{}, 2)

	if !cc.ScheduleAfter(5*time.Millisecond, func() { done <- struct{}{} }) {
		t.Fatal("ScheduleAfter should succeed")
	}
	if !cc.ScheduleAt(time.Now().Add(10*time.Millisecond), func() { done <- struct{}{} }) {
		t.Fatal("ScheduleAt should succeed")
	}

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout waiting exported api task %d", i+1)
		}
	}
}

// TestTaskHandleCancel 验证任务句柄可直接取消已提交任务。
func TestTaskHandleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	var executed atomic.Bool

	h, ok := cc.ScheduleAfterHandle(80*time.Millisecond, func() {
		executed.Store(true)
	})
	if !ok {
		t.Fatal("ScheduleAfterHandle should succeed")
	}
	if h.ID() <= 0 {
		t.Fatalf("invalid handle id: %d", h.ID())
	}
	if !h.Cancel() {
		t.Fatal("handle cancel should succeed")
	}

	if !cc.WaitIdle(10*time.Millisecond, 500*time.Millisecond) {
		t.Fatal("scheduler should become idle after handle cancel")
	}
	if executed.Load() {
		t.Fatal("canceled task must not run")
	}
}

// TestCornCtxCloseRejectsNewTask 验证 Close 后调度器不再接收新任务。
func TestCornCtxCloseRejectsNewTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewCornCtx(ctx)
	cc.Close()

	if cc.ScheduleAfter(5*time.Millisecond, func() {}) {
		t.Fatal("ScheduleAfter should fail after Close")
	}
	if cc.ScheduleAt(time.Now().Add(5*time.Millisecond), func() {}) {
		t.Fatal("ScheduleAt should fail after Close")
	}

	if !cc.WaitDone(500 * time.Millisecond) {
		t.Fatal("scheduler should exit after Close")
	}
}

// TestCornCtxShutdownGracefulAndForced 验证 Shutdown 的优雅关闭和强制关闭语义。
func TestCornCtxShutdownGracefulAndForced(t *testing.T) {
	t.Run("graceful", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cc := NewCornCtx(ctx)
		if !cc.ScheduleAfter(10*time.Millisecond, func() {}) {
			t.Fatal("ScheduleAfter should succeed")
		}

		if !cc.Shutdown(500 * time.Millisecond) {
			t.Fatal("Shutdown should gracefully complete")
		}
		if !cc.WaitDone(0) {
			t.Fatal("scheduler should be done after graceful Shutdown")
		}
	})

	t.Run("forced", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cc := NewCornCtx(ctx)
		start := make(chan struct{}, 1)
		if !cc.ScheduleAfter(1*time.Millisecond, func() {
			start <- struct{}{}
			time.Sleep(120 * time.Millisecond)
		}) {
			t.Fatal("ScheduleAfter should succeed")
		}

		select {
		case <-start:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("task should start before forced shutdown")
		}

		if cc.Shutdown(0) {
			t.Fatal("Shutdown(0) should report forced path")
		}
		if !cc.WaitDone(0) {
			t.Fatal("scheduler should be done after forced Shutdown")
		}
	})
}
