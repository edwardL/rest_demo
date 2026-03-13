package main

import (
	"fmt"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func main1() {
	const (
		targetURL = "http://192.168.50.113:8100/delivery_box/wxapp/rider_recharge/notifyRecharge" // 测试接口
		duration  = 10 * time.Second                                                              // 测试持续时间
		rate      = 1000                                                                          // 每秒请求数
		timeout   = 5 * time.Second                                                               // 请求超时时间
	)

	// 创建攻击目标
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    targetURL,
	})

	// 创建攻击者
	attacker := vegeta.NewAttacker(
		vegeta.Timeout(timeout),
	)

	// 开始压测
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, vegeta.Rate{Freq: rate, Per: time.Second}, duration, "压力测试") {
		metrics.Add(res)
	}
	metrics.Close()

	// 输出结果
	fmt.Printf("\n压测结果:\n")
	fmt.Printf("请求总数: %d\n", metrics.Requests)
	// fmt.Printf("成功请求: %d\n", metrics.success)
	fmt.Printf("失败请求: %v\n", metrics.Errors)
	fmt.Printf("成功率: %.2f%%\n", metrics.Success*100)
	fmt.Printf("QPS: %.2f\n", metrics.Rate)
	fmt.Printf("平均延迟: %v\n", metrics.Latencies.Mean)
	fmt.Printf("P95延迟: %v\n", metrics.Latencies.P95)
	fmt.Printf("P99延迟: %v\n", metrics.Latencies.P99)
	fmt.Printf("最大延迟: %v\n", metrics.Latencies.Max)
}
