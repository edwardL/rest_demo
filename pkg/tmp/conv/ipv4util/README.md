# utils/conv/ipv4util（模型友好版）

`utils/conv/ipv4util` 提供 IPv4 地址处理工具函数。

## 1) 快速使用

```go
package main

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv/ipv4util"
)

func main() {
	// IP 与整数互转
	var ipLong, _ = ipv4util.InetAtoN("192.168.1.1")
	fmt.Println(ipLong) // 3232235777
	var ip = ipv4util.InetNtoA(ipLong)
	fmt.Println(ip) // 192.168.1.1

	// IP 范围解析
	var ip1, ip2, _ = ipv4util.Range("192.168.1.1-192.168.1.10")
	fmt.Println(ip1, ip2)

	// CIDR 解析
	ip1, ip2, _ = ipv4util.Range("192.168.1.0/24")
	fmt.Println(ip2 - ip1 + 1) // 256

	// 判断 IP 是否在范围内
	if ipv4util.Contains("192.168.1.0/24,10.0.0.0/8", "192.168.1.100") {
		fmt.Println("IP 在范围内")
	}

	// 统计 IP 范围数量
	var count = ipv4util.CountIpRange("192.168.1.0/24,10.0.0.0/8")
	fmt.Println(count)

	// 合并连续 IP
	var merged = ipv4util.Merged([]string{"192.168.1.1", "192.168.1.2", "192.168.1.3"})
	fmt.Println(merged) // [192.168.1.1-192.168.1.3]

	// IP 与十六进制互转
	var hex = ipv4util.ToHex("192.168.1.1")
	fmt.Println(hex) // c0a80101
	var ipBack, _ = ipv4util.FromHex("c0a80101")
	fmt.Println(ipBack) // 192.168.1.1
}
```

## 2) 全部导出 API（完整）

### 变量

- `MaskBitMap map[int]string`
    - 掩码位（0~32）与点分十进制掩码的双向映射表

### 函数

- `InetAtoN(ip string) (int64, error)`
    - IPv4 字符串转长整型
- `InetNtoA(ip int64) string`
    - 长整型转 IPv4 字符串
- `Range(ip string) (int64, int64, error)`
    - 解析 IP 范围，支持格式：
        - 单个 IP：`192.168.1.1`
        - 范围：`192.168.1.1-192.168.1.10`
        - CIDR：`192.168.1.0/24`
    - 返回起止 IP 的长整型
- `Contains(ipRangeList string, ip string) bool`
    - 判断 ip 是否在 ipRangeList 中（支持逗号分隔多个范围）
- `CountIpRange(ipRangeList string) int64`
    - 统计 IP 范围内的 IP 数量
- `Merged(ipList []string) []string`
    - 合并连续 IP 为范围表示
- `ToHex(ip string) string`
    - IPv4 转十六进制字符串
- `FromHex(hex string) (string, error)`
    - 十六进制字符串转 IPv4

## 3) 测试命令

```bash
go test ./utils/conv/ipv4util -v
go test ./utils/conv/ipv4util -run '^TestInetAtoNAndInetNtoA$' -v
```
