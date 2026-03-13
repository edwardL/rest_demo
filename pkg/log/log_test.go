package log

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

var logConfig = Config{
	File: "/var/logs//runtime/yema.log",
	//Encoder: "console",
	Level:       "debug",
	Output:      "console",
	Development: true,
}

func TestLogger(t *testing.T) {
	getInt := func(v string) int {
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	}
	str := "| 1024396 | 3251  | 233      | 1720802593 | 354    |\n| 1019263 | 3250  | 225      | 1720799033 | 347    |\n| 1025172 | 3249  | 63       | 1720798991 | 174    |\n| 1015226 | 3248  | 196      | 1720798836 | 317    |\n| 928178  | 3247  | 74       | 1720798718 | 540    |\n| 1021853 | 3246  | 169      | 1720798690 | 288    |\n| 980550  | 3245  | 169      | 1720798641 | 288    |\n| 1022731 | 3244  | 203      | 1720798277 | 326    |\n| 1024861 | 3243  | 233      | 1720796651 | 354    |\n| 1020634 | 3242  | 129      | 1720792630 | 225    |\n| 1024833 | 3241  | 203      | 1720792574 | 326    |\n| 978606  | 3240  | 222      | 1720792077 | 345    |\n| 1024281 | 3239  | 199      | 1720791612 | 321    |\n| 1019392 | 3238  | 167      | 1720791514 | 286    |\n| 1019708 | 3237  | 197      | 1720789077 | 318    |\n| 1024132 | 3236  | 203      | 1720789058 | 326    |\n| 1022936 | 3235  | 196      | 1720787857 | 317    |\n| 1022540 | 3234  | 15       | 1720786979 | 23     |\n| 1023910 | 3233  | 129      | 1720786822 | 225    |\n| 1023892 | 3232  | 233      | 1720786098 | 354    |\n| 1013326 | 3231  | 120      | 1720784974 | 167716 |\n| 1014694 | 3230  | 120      | 1720784672 | 167716 |\n| 1024080 | 3229  | 203      | 1720784407 | 326    |\n| 1019372 | 3228  | 219      | 1720784124 | 342    |\n| 1013326 | 3227  | 120      | 1720784111 | 167716 |\n| 970028  | 3226  | 43       | 1720782403 | 96     |\n| 1013326 | 3225  | 120      | 1720782201 | 167716 |\n| 1020738 | 3224  | 219      | 1720781202 | 342    |\n| 747946  | 3223  | 120      | 1720780825 | 167716 |\n| 957427  | 3222  | 120      | 1720780463 | 167716 |"
	arr := strings.Split(str, "\n")
	for _, s1 := range arr {
		s1 = strings.Trim(s1, "|, ")
		arr1 := strings.Split(s1, "|")
		//fmt.Println(len(arr1))
		sql := "insert into vc_guild_member(id,guild_id,gid,uid,create_by,update_by,leave_type,join_type,created_at,deleted_at)values(%d,%d,%d,%d,5,5,0,2,%d,0);\n"
		fmt.Printf(sql, getInt(arr1[1]), getInt(arr1[2]), getInt(arr1[4]), getInt(arr1[0]), getInt(arr1[3]))
	}
	return
	//str := "1019370 1665411\n1020852 154896\n1020496 52339"
	//arr := strings.Split(str, "\n")
	//mapArr := make(map[int]int)
	//for _, item := range arr {
	//	_arr := strings.SplitN(item, " ", 2)
	//	k := strings.TrimSpace(_arr[0])
	//	k1, _ := strconv.Atoi(k)
	//	v := strings.TrimSpace(_arr[1])
	//	v1, _ := strconv.Atoi(v)
	//	mapArr[k1] = v1
	//}
	//sqlArr := []string{}
	//for uid, coin := range mapArr {
	//	//sql := fmt.Sprintf("select * from  vc_guild_member_salary   where month='2024-07' and uid = %d and status = 1 and coin > %d ;", uid, coin*100)
	//	sql := fmt.Sprintf("update vc_guild_member_salary set coin = coin-%d, coin_bak = %d where month='2024-07' and uid = %d and status = 1 ;", coin*100, coin*100, uid)
	//	sqlArr = append(sqlArr, sql)
	//	fmt.Println(sql)
	//}
	//return
	//fmt.Println(^0)
	joinAt, _ := time.Parse(time.DateTime, "2024-03-12 12:12:12")
	//解析传过来的月份
	month := "2024-03"
	ts, err := time.Parse("2006-01", month)
	if err != nil {
		return
	}
	startTime := ts
	if joinAt.After(ts) {
		startTime = joinAt
	}

	type OnMicDay struct {
		Day        string `json:"day"`
		OnMicTimes int64  `json:"on_mic_times"`
		Coin       int64  `json:"coin"`    //礼物流水
		Diamond    int64  `json:"diamond"` //钻石流水
	}

	//
	_dataList := make([]*OnMicDay, 0)
	i := 0
	now := time.Now()
	for {
		_ts := startTime.Add(time.Hour * 24 * time.Duration(i))
		day := _ts.Format(time.DateOnly)
		i++
		if day[:7] != month || _ts.After(now) {
			break
		}
		_dataList = append(_dataList, &OnMicDay{
			Day: day,
		})
	}

	for _, item := range _dataList {
		fmt.Printf("%+v\n", item)
	}
	//log := NewLog(&logConfig)
	//err := errors.New("test err")
	//log.Debug("test", zap.Int64("id", 43), zap.Error(err))
}

func lottery(arr map[int]int) int {
	cc := int(0)
	for i := range arr {
		cc += arr[i]
	}
	var c int
	for k, v := range arr {
		c = Int(0, cc)
		if c <= v {
			return k
		} else {
			cc -= v
		}
	}
	return 0
}

func lottery2(arr map[int]int) int {
	cc := int(0)
	for i := range arr {
		cc += arr[i]
	}
	n := Int(0, cc)
	vv := 0
	for k, v := range arr {
		vv += v
		if n <= vv {
			return k
		}
	}
	return 0
}
func Int(min, max int) int {
	return rand.Intn(max-min) + min
}

func runTest(n int, arr map[int]int, fn func(arr map[int]int) int) {
	st := time.Now()
	res := make(map[int]int, 0)
	for i := 0; i < n; i++ {
		res[fn(arr)]++
	}
	fmt.Println(res)
	fmt.Println("消耗时间：", time.Now().Sub(st))
}

func TestName(t *testing.T) {
	ts := time.Now()
	fmt.Println(ts)
	fmt.Println(ts.Truncate(time.Hour * 24))
	fmt.Println(ts.Round(time.Hour * 24))
	day := ts.Format(time.DateOnly)
	st, _ := time.Parse(time.DateTime, day+" 00:00:00")
	et, _ := time.Parse(time.DateTime, day+" 23:59:59")
	fmt.Println(day, st, et)
	tz, _ := time.LoadLocation("Asia/Riyadh")
	ts1 := ts.In(tz)
	fmt.Println(ts1)
	fmt.Println(ts1.Truncate(time.Hour * 24))
	fmt.Println(ts1.Round(time.Hour * 24))
	day1 := ts.Format(time.DateOnly)
	st1, _ := time.ParseInLocation(time.DateTime, day1+" 00:00:00", tz)
	et1, _ := time.ParseInLocation(time.DateTime, day1+" 23:59:59", tz)
	fmt.Println(ts1.Location(), day1, st1, et1)
	return
	cc := 333
	str := fmt.Sprintf("%d", cc)
	_b := []byte(str)
	flag := false
	for i := 0; i < len(_b)-1; i++ {
		if _b[i] != _b[i+1] {
			flag = true
			break
		}
	}
	if !flag {
		cc++
		str = fmt.Sprintf("%d", cc)
	}
	fmt.Println(str)
	return

	var a = map[int]int{
		3: 45000,
		2: 10000,
		4: 35000,
		1: 10000,
	}

	n := 1000000
	runTest(n, a, lottery)
	runTest(n, a, lottery2)
}
