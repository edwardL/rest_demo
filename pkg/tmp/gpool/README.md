# gpool

通用、可伸缩、可观测的 goroutine 池。

## 功能特性

- 动态伸缩：根据队列压力扩容，空闲超时后缩容到 `MinWorkers`
- 可靠执行：任务 panic 会被 recover，不影响池继续运行
- 生命周期管理：`Close()` + `Shutdown(timeout)`
- 可观测统计：`Stats()` 返回队列、运行中、完成数、拒绝数、panic 数等
- 同步提交：`SubmitWait(ctx, fn)` 支持“提交并等待执行完成”
- 非阻塞尝试提交：`SubmitTry(ctx, fn)`
- 队列策略：支持 `Block` / `Reject` / `CallerRuns`
- 预设配置：默认、零队列、非阻塞拒绝、突发流量友好

## 快速使用

```go
opts := gpool.DefaultOptions()
opts.MaxWorkers = 32
opts.MinWorkers = 2

pool, err := gpool.New(context.Background(), opts)
if err != nil {
    panic(err)
}
defer pool.Shutdown(2 * time.Second)

_ = pool.Submit(context.Background(), func() {
    // async task
})

_ = pool.SubmitWait(context.Background(), func() {
    // wait until done
})
```

## API 说明

- `New(ctx, opts)`：创建池
- `DefaultOptions()`：默认预设，适合通用异步任务
- `ZeroQueueOptions()`：零队列预设，适合快速背压
- `NonBlockingOptions()`：非阻塞拒绝预设，适合更关注调用方时延
- `BurstOptions()`：突发友好预设，适合轻任务高吞吐场景
- `Submit(ctx, fn)`：异步提交
- `SubmitTry(ctx, fn)`：尝试立即提交，失败返回 `ErrQueueFull`
- `SubmitWait(ctx, fn)`：提交并等待任务执行结束
- `WaitIdle(timeout)`：等待池空闲（queued=0 且 running=0）
- `Close()`：停止接收新任务
- `Shutdown(timeout)`：优雅/强制关闭
- `Stats()`：获取统计快照

### PanicHandler

可通过 `Options.PanicHandler` 注入 panic 处理逻辑：

```go
opts := gpool.DefaultOptions()
opts.PanicHandler = func(v any) {
    // log / metrics / report
}
```

## 队列策略说明

### QueuePolicyBlock

- 队列满时阻塞等待
- 适合希望通过背压保护系统的场景

### QueuePolicyReject

- 队列满时立即返回 `ErrQueueFull`
- 适合不能接受阻塞的实时请求链路

### QueuePolicyCallerRuns

- 队列满时由调用方直接执行任务
- 适合需要自然限流、让上游自己承担压力的场景

## 如何选择预设

- `DefaultOptions()`：通用默认值，`QueueSize = MaxWorkers * 8`
- `ZeroQueueOptions()`：无排队缓冲，尽快暴露压力
- `NonBlockingOptions()`：有界队列 + 满即拒绝
- `BurstOptions()`：较大队列，适合吸收短时流量尖峰

## 测试与性能

### 单元测试

```bash
GO111MODULE=off go test ./utils/gpool
```

### 基准测试

```bash
GO111MODULE=off go test ./utils/gpool -bench BenchmarkPool -run ^$
```

### 性能画像测试（默认关闭）

```bash
HHIT_RUN_PERF_TESTS=1 GO111MODULE=off go test ./utils/gpool -run TestPoolLatencyProfile
```
