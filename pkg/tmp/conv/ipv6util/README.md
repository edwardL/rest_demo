# utils/conv/ipv6util（模型友好版）

`utils/conv/ipv6util` 提供 IPv6 地址处理工具函数。

## 1) 快速使用

```go
package main

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv/ipv6util"
)

func main() {
	// IPv6 与大整数互转
	var ipLong, _ = ipv6util.InetAtoN("2001:db8::1")
	fmt.Println(ipLong.String())
	var ip = ipv6util.InetNtoA(ipLong)
	fmt.Println(ip)

	// IPv6 范围解析
	var ip1, ip2, _ = ipv6util.Range("2001:db8::1-2001:db8::10")
	fmt.Println(ip1, ip2)

	// CIDR 解析
	ip1, ip2, _ = ipv6util.Range("2001:db8::/32")
	var diff = ip2.Sub(ip2, ip1)
	fmt.Println("IP数量:", diff.Add(diff, big.NewInt(1)))
}
```

## 2) 全部导出 API（完整）

### 变量

- `MaskBitMap map[int]string`
    - 掩码位（1~128）与 IPv6 掩码的双向映射表

### 函数

- `InetAtoN(ip string) (*big.Int, error)`
    - IPv6 字符串转 `*big.Int`
- `InetNtoA(ip *big.Int) string`
    - `*big.Int` 转 IPv6 字符串
- `Range(ip string) (*big.Int, *big.Int, error)`
    - 解析 IPv6 范围，支持格式：
        - 单个 IPv6：`2001:db8::1`
        - 范围：`2001:db8::1-2001:db8::10`
        - CIDR：`2001:db8::/32`
    - 返回起止 IPv6 的大整数表示

## 3) 注意事项

- IPv6 地址空间巨大，使用 `*big.Int` 表示
- 范围计算可能产生超大数值，注意内存使用
- CIDR 掩码位范围 1~128

## 4) 测试命令

```bash
go test ./utils/conv/ipv6util -v
go test ./utils/conv/ipv6util -run '^TestInetAtoNAndInetNtoA$' -v
```
