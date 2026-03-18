package ipv6util

import (
	"math/big"
	"testing"
)

func TestInetAtoNAndInetNtoA(t *testing.T) {
	var ip = "2001:db8::1"
	var ipLong *big.Int
	var err error
	ipLong, err = InetAtoN(ip)
	if err != nil {
		t.Fatalf("InetAtoN 失败: %v", err)
	}

	var ipBack string
	ipBack = InetNtoA(ipLong)
	// 由于 IPv6 表示有多种形式，这里只验证能解析回来
	var ipBackLong *big.Int
	ipBackLong, err = InetAtoN(ipBack)
	if err != nil {
		t.Fatalf("InetNtoA 结果无法解析: %v", err)
	}
	if ipLong.Cmp(ipBackLong) != 0 {
		t.Fatalf("InetAtoN/InetNtoA 往返不一致: %s != %s", ipLong.String(), ipBackLong.String())
	}

	_, err = InetAtoN("invalid")
	if err == nil {
		t.Fatalf("InetAtoN 应拒绝非法 IP")
	}
}

func TestRange(t *testing.T) {
	var ip1, ip2 *big.Int
	var err error

	// 单个 IP
	ip1, ip2, err = Range("2001:db8::1")
	if err != nil || ip1.Cmp(ip2) != 0 {
		t.Fatalf("Range 单IP错误: %v", err)
	}

	// 范围格式
	ip1, ip2, err = Range("2001:db8::1-2001:db8::10")
	if err != nil {
		t.Fatalf("Range 范围格式失败: %v", err)
	}
	var diff = big.NewInt(0).Sub(ip2, ip1)
	if diff.Int64() != 15 {
		t.Fatalf("Range 范围差值错误: %d", diff.Int64())
	}

	// CIDR 格式
	ip1, ip2, err = Range("2001:db8::/32")
	if err != nil {
		t.Fatalf("Range CIDR 失败: %v", err)
	}
	diff = big.NewInt(0).Sub(ip2, ip1)
	// /32 表示前32位固定，后96位可变，2^96 个IP
	var expectedDiff = big.NewInt(1)
	expectedDiff.Lsh(expectedDiff, 96)
	expectedDiff.Sub(expectedDiff, big.NewInt(1))
	if diff.Cmp(expectedDiff) != 0 {
		t.Fatalf("Range /32 范围错误: %s != %s", diff.String(), expectedDiff.String())
	}
}

func TestMaskBitMap(t *testing.T) {
	if MaskBitMap[32] != "FFFF:FFFF::" {
		t.Fatalf("MaskBitMap[32] 错误: %s", MaskBitMap[32])
	}
	if MaskBitMap[64] != "FFFF:FFFF:FFFF:FFFF::" {
		t.Fatalf("MaskBitMap[64] 错误: %s", MaskBitMap[64])
	}
	if MaskBitMap[128] != "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF" {
		t.Fatalf("MaskBitMap[128] 错误: %s", MaskBitMap[128])
	}
	if len(MaskBitMap) != 128 {
		t.Fatalf("MaskBitMap 长度应为128: %d", len(MaskBitMap))
	}
}

func TestIPv6SpecialCases(t *testing.T) {
	var testCases = []struct {
		name string
		ip   string
	}{
		{"全零", "::"},
		{"回环", "::1"},
		{"全一", "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF"},
		{"文档前缀", "2001:db8::1"},
		{"链路本地", "fe80::1"},
	}

	var i int
	for i = 0; i < len(testCases); i++ {
		var tc = testCases[i]
		var ipLong *big.Int
		var err error
		ipLong, err = InetAtoN(tc.ip)
		if err != nil {
			t.Fatalf("%s InetAtoN 失败: %v", tc.name, err)
		}
		var ipBack = InetNtoA(ipLong)
		var ipBackLong *big.Int
		ipBackLong, err = InetAtoN(ipBack)
		if err != nil {
			t.Fatalf("%s InetNtoA 结果无法解析: %v", tc.name, err)
		}
		if ipLong.Cmp(ipBackLong) != 0 {
			t.Fatalf("%s 往返不一致", tc.name)
		}
	}
}
