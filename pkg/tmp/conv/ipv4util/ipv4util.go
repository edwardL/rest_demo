package ipv4util

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

// MaskBitMap 掩码位与掩码的点分十进制的双向对应关系
var MaskBitMap = map[int]string{
	0:  "0.0.0.0",
	1:  "128.0.0.0",
	2:  "192.0.0.0",
	3:  "224.0.0.0",
	4:  "240.0.0.0",
	5:  "248.0.0.0",
	6:  "252.0.0.0",
	7:  "254.0.0.0",
	8:  "255.0.0.0",
	9:  "255.128.0.0",
	10: "255.192.0.0",
	11: "255.224.0.0",
	12: "255.240.0.0",
	13: "255.248.0.0",
	14: "255.252.0.0",
	15: "255.254.0.0",
	16: "255.255.0.0",
	17: "255.255.128.0",
	18: "255.255.192.0",
	19: "255.255.224.0",
	20: "255.255.240.0",
	21: "255.255.248.0",
	22: "255.255.252.0",
	23: "255.255.254.0",
	24: "255.255.255.0",
	25: "255.255.255.128",
	26: "255.255.255.192",
	27: "255.255.255.224",
	28: "255.255.255.240",
	29: "255.255.255.248",
	30: "255.255.255.252",
	31: "255.255.255.254",
	32: "255.255.255.255",
}

// InetAtoN 转换IP地址字符串为长整型
func InetAtoN(ip string) (int64, error) {
	tmpS := net.ParseIP(ip)
	if tmpS == nil {
		return 0, fmt.Errorf("illegal ip %s", ip)
	}
	bits := strings.Split(ip, ".")
	if len(bits) != 4 {
		return 0, fmt.Errorf("illegal ip %s", ip)
	}
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum, nil
}

// InetNtoA 转换整型IP地址为常规地址字符串
func InetNtoA(ip int64) string {
	var tmpBytes [4]byte
	tmpBytes[0] = byte(ip & 0xFF)
	tmpBytes[1] = byte((ip >> 8) & 0xFF)
	tmpBytes[2] = byte((ip >> 16) & 0xFF)
	tmpBytes[3] = byte((ip >> 24) & 0xFF)
	return net.IPv4(tmpBytes[3], tmpBytes[2], tmpBytes[1], tmpBytes[0]).String()
}

// Range 获取IP地址的范围,支持的格式: 192.168.11.50, 192.168.11.50/26, 192.168.11.50-192.168.11.60
func Range(ip string) (int64, int64, error) {
	if strings.Contains(ip, "-") {
		ips := strings.Split(ip, "-")
		ip1, err := InetAtoN(ips[0])
		if err != nil {
			return 0, 0, err
		}
		ip2, err := InetAtoN(ips[1])
		if err != nil {
			return 0, 0, err
		}
		return ip1, ip2, nil
	}
	if strings.Contains(ip, "/") {
		ips := strings.Split(ip, "/")
		ipN, err := InetAtoN(ips[0])
		if err != nil {
			return 0, 0, err
		}
		mask, err := strconv.Atoi(ips[1])
		if err != nil {
			return 0, 0, err
		}
		ipMask, err := InetAtoN(MaskBitMap[mask])
		if err != nil {
			return 0, 0, err
		}
		ip1 := ipN & ipMask
		ip2, err := InetAtoN(InetNtoA(ip1 + ^ipMask))
		return ip1, ip2, err
	}
	p0, err := InetAtoN(ip)
	return p0, p0, err
}

// Contains 判断目标IP是否在IP范围内
func Contains(ipRangeList string, ip string) bool {
	if len(ipRangeList) == 0 || len(ip) == 0 {
		return false
	}
	ipLong, err := InetAtoN(ip)
	if err != nil {
		return false
	}
	for _, ipRange := range strings.Split(ipRangeList, ",") {
		ip1, ip2, err := Range(ipRange)
		if err != nil {
			return false
		}
		if ip1 <= ipLong && ipLong <= ip2 {
			return true
		}
	}
	return false
}

// CountIpRange 统计IP范围的数据
func CountIpRange(ipRangeList string) int64 {
	var result int64
	for _, ipRang := range strings.Split(ipRangeList, ",") {
		ip1, ip2, err := Range(ipRang)
		if err != nil {
			return 0
		}
		result += ip2 - ip1 + 1
	}
	return result
}

// Merged 合并IP
func Merged(ipList []string) []string {
	var err error
	var ip string
	var ipLong int64
	var i int
	var ipLongList = make([]int64, len(ipList))
	for i, ip = range ipList {
		ipLong, err = InetAtoN(ip)
		if err != nil {
			continue
		}
		ipLongList[i] = ipLong
	}
	sort.Slice(ipLongList, func(i, j int) bool {
		return ipLongList[i] < ipLongList[j]
	})
	var merged = [][]int64{{ipLongList[0], ipLongList[0]}}
	for i, ipLong = range ipLongList[1:] {
		if ipLong-merged[len(merged)-1][1] <= 1 {
			merged[len(merged)-1][1] = ipLong
		} else {
			merged = append(merged, []int64{ipLong, ipLong})
		}
	}
	var result = make([]string, len(merged))
	for i, item := range merged {
		if item[0] == item[1] {
			result[i] = InetNtoA(item[0])
		} else {
			result[i] = fmt.Sprintf("%s-%s", InetNtoA(item[0]), InetNtoA(item[1]))
		}
	}
	return result
}

// ToHex IP转换为16进制
func ToHex(ip string) string {
	var parseIP = net.ParseIP(ip)
	if parseIP == nil {
		return ""
	}
	return fmt.Sprintf("%02x%02x%02x%02x", parseIP[12], parseIP[13], parseIP[14], parseIP[15])
}

// FromHex 16进制转换为IP
func FromHex(hex string) (string, error) {
	var err error
	var ip strings.Builder
	var decimal int64
	decimal, err = strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return "", err
	}
	ip.WriteString(strconv.FormatInt(decimal, 10))
	ip.WriteString(".")
	decimal, err = strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return "", err
	}
	ip.WriteString(strconv.FormatInt(decimal, 10))
	ip.WriteString(".")
	decimal, err = strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return "", err
	}
	ip.WriteString(strconv.FormatInt(decimal, 10))
	ip.WriteString(".")
	decimal, err = strconv.ParseInt(hex[6:8], 16, 64)
	if err != nil {
		return "", err
	}
	ip.WriteString(strconv.FormatInt(decimal, 10))
	return ip.String(), nil
}
