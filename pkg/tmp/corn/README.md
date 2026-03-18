# corn

轻量级内存调度器（按到期时间执行任务）。

## 功能特性

- 支持绝对时间与相对时间调度：`ScheduleAt` / `ScheduleAfter`
- 支持任务取消：`Cancel(id)` 或 `TaskHandle.Cancel()`
- 支持等待语义：`WaitAll` / `WaitIdle` / `WaitDone`
- 支持关闭语义：`Close()` / `Shutdown(timeout)`
- 单线程调度执行，按到期顺序出队执行

## 快速使用

```go
ctx := context.Background()
cc := corn.NewCornCtx(ctx)

h, ok := cc.ScheduleAfterHandle(200*time.Millisecond, func() {
    // do work
})
if ok {
    _ = h.Cancel() // optional
}

_ = cc.ScheduleAt(time.Now().Add(time.Second), func() {
    // run at specific time
})

_ = cc.WaitIdle(10*time.Millisecond, 2*time.Second)
_ = cc.Shutdown(2 * time.Second)
```

## API 说明

- 调度：
  - `ScheduleAt(t, fn)` / `ScheduleAtID` / `ScheduleAtHandle`
  - `ScheduleAfter(td, fn)` / `ScheduleAfterID` / `ScheduleAfterHandle`
- 取消：
  - `Cancel(id)`
  - `TaskHandle.Cancel()`
- 等待：
  - `WaitAll(timeout)`：等待所有已接收任务完成
  - `WaitIdle(idle, timeout)`：等待空闲窗口稳定
  - `WaitDone(timeout)`：等待调度器 goroutine 退出
- 关闭：
  - `Close()`：立即停止接收新任务
  - `Shutdown(timeout)`：优雅/强制关闭

## 测试与性能

### 单元测试

```bash
GO111MODULE=off go test ./utils/corn
```

### 基准测试

```bash
GO111MODULE=off go test ./utils/corn -bench BenchmarkCornCtx -run ^$
```

### 性能画像测试（默认关闭）

```bash
HHIT_RUN_PERF_TESTS=1 GO111MODULE=off go test ./utils/corn -run 'TestCornCtx(TimingPrecisionProfile|ConcurrentSubmitPrecision|ConcurrentTailLatencyLongRun)'
```
