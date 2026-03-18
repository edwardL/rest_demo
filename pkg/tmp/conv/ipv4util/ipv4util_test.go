package ipv4util

import (
	"testing"
)

func TestInetAtoNAndInetNtoA(t *testing.T) {
	var ip = "192.168.1.1"
	var ipLong int64
	var err error
	ipLong, err = InetAtoN(ip)
	if err != nil {
		t.Fatalf("InetAtoN 失败: %v", err)
	}

	var ipBack string
	ipBack = InetNtoA(ipLong)
	if ipBack != ip {
		t.Fatalf("InetNtoA 结果错误: %s != %s", ipBack, ip)
	}

	_, err = InetAtoN("invalid")
	if err == nil {
		t.Fatalf("InetAtoN 应拒绝非法 IP")
	}

	_, err = InetAtoN("192.168.1")
	if err == nil {
		t.Fatalf("InetAtoN 应拒绝非 4 段 IP")
	}
}

func TestRange(t *testing.T) {
	var ip1, ip2 int64
	var err error

	// 单个 IP
	ip1, ip2, err = Range("192.168.1.1")
	if err != nil || ip1 != ip2 {
		t.Fatalf("Range 单IP错误: %d %d %v", ip1, ip2, err)
	}

	// 范围格式
	ip1, ip2, err = Range("192.168.1.1-192.168.1.10")
	if err != nil || ip2-ip1 != 9 {
		t.Fatalf("Range 范围格式错误: %d %d %v", ip1, ip2, err)
	}

	// CIDR 格式
	ip1, ip2, err = Range("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Range CIDR 失败: %v", err)
	}
	if ip2-ip1+1 != 256 {
		t.Fatalf("Range /24 应有256个IP: %d", ip2-ip1+1)
	}
}

func TestContains(t *testing.T) {
	if !Contains("192.168.1.1-192.168.1.10", "192.168.1.5") {
		t.Fatalf("Contains 应包含范围中的 IP")
	}
	if Contains("192.168.1.1-192.168.1.10", "192.168.1.20") {
		t.Fatalf("Contains 不应包含范围外的 IP")
	}
	if !Contains("192.168.1.0/24", "192.168.1.100") {
		t.Fatalf("Contains 应包含 CIDR 中的 IP")
	}
	if Contains("", "192.168.1.1") {
		t.Fatalf("Contains 空范围应返回 false")
	}
	if Contains("192.168.1.1", "") {
		t.Fatalf("Contains 空 IP 应返回 false")
	}
	if Contains("192.168.1.1-192.168.1.10", "invalid") {
		t.Fatalf("Contains 非法 IP 应返回 false")
	}
}

func TestCountIpRange(t *testing.T) {
	var count = CountIpRange("192.168.1.1-192.168.1.10")
	if count != 10 {
		t.Fatalf("CountIpRange 单范围错误: %d", count)
	}

	count = CountIpRange("192.168.1.0/24")
	if count != 256 {
		t.Fatalf("CountIpRange CIDR 错误: %d", count)
	}

	count = CountIpRange("192.168.1.1-192.168.1.5,192.168.2.1-192.168.2.3")
	if count != 8 {
		t.Fatalf("CountIpRange 多范围错误: %d", count)
	}

	count = CountIpRange("invalid")
	if count != 0 {
		t.Fatalf("CountIpRange 非法格式应返回0: %d", count)
	}
}

func TestMerged(t *testing.T) {
	var ipList = []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.10"}
	var merged = Merged(ipList)
	if len(merged) != 2 {
		t.Fatalf("Merged 应合并连续IP: %v", merged)
	}
	if merged[0] != "192.168.1.1-192.168.1.3" {
		t.Fatalf("Merged 第一个范围错误: %s", merged[0])
	}
	if merged[1] != "192.168.1.10" {
		t.Fatalf("Merged 第二个范围错误: %s", merged[1])
	}
}

func TestToHexAndFromHex(t *testing.T) {
	var ip = "192.168.1.1"
	var hex = ToHex(ip)
	if hex != "c0a80101" {
		t.Fatalf("ToHex 结果错误: %s", hex)
	}

	var ipBack string
	var err error
	ipBack, err = FromHex(hex)
	if err != nil || ipBack != ip {
		t.Fatalf("FromHex 结果错误: %s %v", ipBack, err)
	}

	hex = ToHex("invalid")
	if hex != "" {
		t.Fatalf("ToHex 非法IP应返回空: %s", hex)
	}

	_, err = FromHex("invalid")
	if err == nil {
		t.Fatalf("FromHex 非法hex应返回错误")
	}
}

func TestMaskBitMap(t *testing.T) {
	if MaskBitMap[24] != "255.255.255.0" {
		t.Fatalf("MaskBitMap[24] 错误: %s", MaskBitMap[24])
	}
	if MaskBitMap[16] != "255.255.0.0" {
		t.Fatalf("MaskBitMap[16] 错误: %s", MaskBitMap[16])
	}
	if len(MaskBitMap) != 33 {
		t.Fatalf("MaskBitMap 长度应为33: %d", len(MaskBitMap))
	}
}

func TestIPv4AllFunctions(t *testing.T) {
	var testCases = []struct {
		name string
		ip   string
	}{
		{"0.0.0.0", "0.0.0.0"},
		{"127.0.0.1", "127.0.0.1"},
		{"10.0.0.1", "10.0.0.1"},
		{"172.16.0.1", "172.16.0.1"},
		{"192.168.100.200", "192.168.100.200"},
		{"255.255.255.255", "255.255.255.255"},
	}

	var i int
	for i = 0; i < len(testCases); i++ {
		var tc = testCases[i]
		var ipLong int64
		var err error
		ipLong, err = InetAtoN(tc.ip)
		if err != nil {
			t.Fatalf("%s InetAtoN 失败: %v", tc.name, err)
		}
		var ipBack = InetNtoA(ipLong)
		if ipBack != tc.ip {
			t.Fatalf("%s 往返错误: %s != %s", tc.name, ipBack, tc.ip)
		}
	}
}
