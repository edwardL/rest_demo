package ipv6util

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
)

// MaskBitMap 掩码位与掩码的点分十进制的双向对应关系
var MaskBitMap = map[int]string{
	1:   "8000::",
	2:   "C000::",
	3:   "E000::",
	4:   "F000::",
	5:   "F800::",
	6:   "FC00::",
	7:   "FE00::",
	8:   "FF00::",
	9:   "FF80::",
	10:  "FFC0::",
	11:  "FFE0::",
	12:  "FFF0::",
	13:  "FFF8::",
	14:  "FFFC::",
	15:  "FFFE::",
	16:  "FFFF::",
	17:  "FFFF:8000::",
	18:  "FFFF:C000::",
	19:  "FFFF:E000::",
	20:  "FFFF:F000::",
	21:  "FFFF:F800::",
	22:  "FFFF:FC00::",
	23:  "FFFF:FE00::",
	24:  "FFFF:FF00::",
	25:  "FFFF:FF80::",
	26:  "FFFF:FFC0::",
	27:  "FFFF:FFE0::",
	28:  "FFFF:FFF0::",
	29:  "FFFF:FFF8::",
	30:  "FFFF:FFFC::",
	31:  "FFFF:FFFE::",
	32:  "FFFF:FFFF::",
	33:  "FFFF:FFFF:8000::",
	34:  "FFFF:FFFF:C000::",
	35:  "FFFF:FFFF:E000::",
	36:  "FFFF:FFFF:F000::",
	37:  "FFFF:FFFF:F800::",
	38:  "FFFF:FFFF:FC00::",
	39:  "FFFF:FFFF:FE00::",
	40:  "FFFF:FFFF:FF00::",
	41:  "FFFF:FFFF:FF80::",
	42:  "FFFF:FFFF:FFC0::",
	43:  "FFFF:FFFF:FFE0::",
	44:  "FFFF:FFFF:FFF0::",
	45:  "FFFF:FFFF:FFF8::",
	46:  "FFFF:FFFF:FFFC::",
	47:  "FFFF:FFFF:FFFE::",
	48:  "FFFF:FFFF:FFFF::",
	49:  "FFFF:FFFF:FFFF:8000::",
	50:  "FFFF:FFFF:FFFF:C000::",
	51:  "FFFF:FFFF:FFFF:E000::",
	52:  "FFFF:FFFF:FFFF:F000::",
	53:  "FFFF:FFFF:FFFF:F800::",
	54:  "FFFF:FFFF:FFFF:FC00::",
	55:  "FFFF:FFFF:FFFF:FE00::",
	56:  "FFFF:FFFF:FFFF:FF00::",
	57:  "FFFF:FFFF:FFFF:FF80::",
	58:  "FFFF:FFFF:FFFF:FFC0::",
	59:  "FFFF:FFFF:FFFF:FFE0::",
	60:  "FFFF:FFFF:FFFF:FFF0::",
	61:  "FFFF:FFFF:FFFF:FFF8::",
	62:  "FFFF:FFFF:FFFF:FFFC::",
	63:  "FFFF:FFFF:FFFF:FFFE::",
	64:  "FFFF:FFFF:FFFF:FFFF::",
	65:  "FFFF:FFFF:FFFF:FFFF:8000::",
	66:  "FFFF:FFFF:FFFF:FFFF:C000::",
	67:  "FFFF:FFFF:FFFF:FFFF:E000::",
	68:  "FFFF:FFFF:FFFF:FFFF:F000::",
	69:  "FFFF:FFFF:FFFF:FFFF:F800::",
	70:  "FFFF:FFFF:FFFF:FFFF:FC00::",
	71:  "FFFF:FFFF:FFFF:FFFF:FE00::",
	72:  "FFFF:FFFF:FFFF:FFFF:FF00::",
	73:  "FFFF:FFFF:FFFF:FFFF:FF80::",
	74:  "FFFF:FFFF:FFFF:FFFF:FFC0::",
	75:  "FFFF:FFFF:FFFF:FFFF:FFE0::",
	76:  "FFFF:FFFF:FFFF:FFFF:FFF0::",
	77:  "FFFF:FFFF:FFFF:FFFF:FFF8::",
	78:  "FFFF:FFFF:FFFF:FFFF:FFFC::",
	79:  "FFFF:FFFF:FFFF:FFFF:FFFE::",
	80:  "FFFF:FFFF:FFFF:FFFF:FFFF::",
	81:  "FFFF:FFFF:FFFF:FFFF:FFFF:8000::",
	82:  "FFFF:FFFF:FFFF:FFFF:FFFF:C000::",
	83:  "FFFF:FFFF:FFFF:FFFF:FFFF:E000::",
	84:  "FFFF:FFFF:FFFF:FFFF:FFFF:F000::",
	85:  "FFFF:FFFF:FFFF:FFFF:FFFF:F800::",
	86:  "FFFF:FFFF:FFFF:FFFF:FFFF:FC00::",
	87:  "FFFF:FFFF:FFFF:FFFF:FFFF:FE00::",
	88:  "FFFF:FFFF:FFFF:FFFF:FFFF:FF00::",
	89:  "FFFF:FFFF:FFFF:FFFF:FFFF:FF80::",
	90:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFC0::",
	91:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFE0::",
	92:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFF0::",
	93:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFF8::",
	94:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFC::",
	95:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFE::",
	96:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF::",
	97:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:8000:0000",
	98:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:C000:0000",
	99:  "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:E000:0000",
	100: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:F000:0000",
	101: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:F800:0000",
	102: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FC00:0000",
	103: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FE00:0000",
	104: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FF00:0000",
	105: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FF80:0000",
	106: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFC0:0000",
	107: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFE0:0000",
	108: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFF0:0000",
	109: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFF8:0000",
	110: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFC:0000",
	111: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFE:0000",
	112: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:0000",
	113: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:8000",
	114: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:C000",
	115: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:E000",
	116: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:F000",
	117: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:F800",
	118: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FC00",
	119: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FE00",
	120: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FF00",
	121: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FF80",
	122: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFC0",
	123: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFE0",
	124: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFF0",
	125: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFF8",
	126: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFC",
	127: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFE",
	128: "FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF",
}

// InetAtoN 转换IP地址字符串为长整型
func InetAtoN(ip string) (*big.Int, error) {
	tmpS := net.ParseIP(ip)
	if tmpS == nil {
		return big.NewInt(0), fmt.Errorf("illegal ip %s", ip)
	}
	result := big.NewInt(0)
	result.SetBytes(tmpS)
	return result, nil
}

// InetNtoA 转换整型IP地址为常规地址字符串
func InetNtoA(ip *big.Int) string {
	b255 := new(big.Int).SetBytes([]byte{255})
	var buf = make([]byte, 2)
	p := make([]string, 8)
	j := 0
	var i uint
	tmpInt := new(big.Int)
	for i = 0; i < 16; i += 2 {
		tmpInt.Rsh(ip, 120-i*8).And(tmpInt, b255)
		bytes := tmpInt.Bytes()
		if len(bytes) > 0 {
			buf[0] = bytes[0]
		} else {
			buf[0] = 0
		}
		tmpInt.Rsh(ip, 120-(i+1)*8).And(tmpInt, b255)
		bytes = tmpInt.Bytes()
		if len(bytes) > 0 {
			buf[1] = bytes[0]
		} else {
			buf[1] = 0
		}
		p[j] = hex.EncodeToString(buf)
		j++
	}
	return strings.Join(p, ":")
}

// Range 获取IP地址的范围,支持的格式: 0:0:0:0:0:0:0:0, 0:0:0:0:0:0:0:0/26, 0:0:0:0:0:0:0:0-0:0:0:0:0:0:0:60
func Range(ip string) (*big.Int, *big.Int, error) {
	if strings.Contains(ip, "-") {
		ips := strings.Split(ip, "-")
		ip1, err := InetAtoN(ips[0])
		if err != nil {
			return big.NewInt(0), big.NewInt(0), err
		}
		ip2, err := InetAtoN(ips[1])
		if err != nil {
			return big.NewInt(0), big.NewInt(0), err
		}
		return ip1, ip2, nil
	}
	if strings.Contains(ip, "/") {
		ips := strings.Split(ip, "/")
		ipN, err := InetAtoN(ips[0])
		if err != nil {
			return big.NewInt(0), big.NewInt(0), err
		}
		mask, err := strconv.Atoi(ips[1])
		if err != nil {
			return big.NewInt(0), big.NewInt(0), err
		}
		ipMask, err := InetAtoN(MaskBitMap[mask])
		if err != nil {
			return big.NewInt(0), big.NewInt(0), err
		}
		ip1 := ipN.And(ipN, ipMask)
		add := big.NewInt(0).Add(ip1, ipMask.Not(ipMask))
		ip2, err := InetAtoN(InetNtoA(add))
		return ip1, ip2, err
	}
	p0, err := InetAtoN(ip)
	return p0, p0, err
}
