package gdbtmp

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// nil 值会忽略这个字段

// dbModel db模型
type dbModel struct {
	Id          int64     `json:"id"`                                                 // ID
	Ts          int64     `json:"ts"`                                                 // TS
	Uid         string    `json:"uid"`                                                // uid
	Ip          string    `json:"ip"`                                                 // IP
	HostIp      string    `json:"host_ip"`                                            // 主机IP
	TestTime    time.Time `json:"test_time" gdbtmp:"ip_long;type:Unix"`               // 测试时间 使用gdb重命名和设置时间格式为秒
	TestTimeMil time.Time `json:"test_time_mil" gdbtmp:"peer_ip_long;type:UnixMilli"` // 测试时间 使用gdb重命名和设置时间格式为毫秒
	CreateId    int64     `json:"create_id"`                                          // 创建者ID
	CreateTs    int64     `json:"create_ts"`                                          // 创建者TS
	CreateTime  time.Time `json:"create_time" gdbtmp:"type:2006-01-02"`               // 创建时间 使用gdb设置时间格式为 年-月-日
	UpdateId    int64     `json:"update_id"`                                          // 更新者ID
	UpdateTs    int64     `json:"update_ts"`                                          // 更新者TS
	UpdateTime  time.Time `json:"update_time"`                                        // 更新时间 time.Time类型默认时间格式为 年-月-日 时:分:秒
}

// TableName 表名称
func (*dbModel) TableName() string {
	return "`hhit_asset`.`hhit_agent_network`"
}

// dbModel db模型
type dbModelPtr struct {
	Id          *int64     `json:"id"`                                                 // ID
	Ts          *int64     `json:"ts"`                                                 // TS
	Uid         *string    `json:"uid"`                                                // uid
	Ip          *string    `json:"ip"`                                                 // IP
	HostIp      *string    `json:"host_ip"`                                            // 主机IP
	TestTime    *time.Time `json:"test_time" gdbtmp:"ip_long;type:Unix"`               // 测试时间 使用gdb重命名和设置时间格式为秒
	TestTimeMil *time.Time `json:"test_time_mil" gdbtmp:"peer_ip_long;type:UnixMilli"` // 测试时间 使用gdb重命名和设置时间格式为毫秒
	CreateId    *int64     `json:"create_id"`                                          // 创建者ID
	CreateTs    *int64     `json:"create_ts"`                                          // 创建者TS
	CreateTime  *time.Time `json:"create_time" gdbtmp:"type:2006-01-02"`               // 创建时间 使用gdb设置时间格式为 年-月-日
	UpdateId    *int64     `json:"update_id"`                                          // 更新者ID
	UpdateTs    *int64     `json:"update_ts"`                                          // 更新者TS
	UpdateTime  *time.Time `json:"update_time"`                                        // 更新时间 time.Time类型默认时间格式为 年-月-日 时:分:秒
}

// TableName 表名称
func (*dbModelPtr) TableName() string {
	return "`hhit_asset`.`hhit_agent_network`"
}

func dbInit() {
	initLog()
	var err error
	var dbConf = map[string]string{
		"db_host":            "",
		"db_port":            "",
		"db_user":            "",
		"db_pass":            "",
		"db_name":            "",
		"db_charset":         "utf8mb4",
		"db_type":            "mysql",
		"need_mode_onebyone": "1",
	}
	//err = hhdb.InitDB(dbConf["db_type"], &dbConf)
	if err != nil {
		fmt.Println(err)
	}
}

func TestDb_Scan(t *testing.T) {
	dbInit()
	var dbData dbModel
	var err = New(&dbData).Where("id = ?", 482000699).
		Scan(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbData)

	var dbDataList []*dbModel
	err = New(&dbModel{}).Limit("?", 3).
		Scan(&dbDataList).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbDataList)

	m, err := New(&dbModel{}).Limit("?", 3).
		Maps()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(m)
}

func TestDb_ScanPtr(t *testing.T) {
	dbInit()
	var dbData dbModelPtr
	var err = New(&dbData).Where("id = ?", 482000699).
		Scan(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbData)

	var dbDataList []*dbModelPtr
	err = New(&dbModel{}).Limit("?", 3).
		Scan(&dbDataList).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbDataList)

	m, err := New(&dbModel{}).Limit("?", 3).
		Maps()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(m)
}

func TestDb_ScanMap(t *testing.T) {
	dbInit()
	var dbData map[string]any
	var err = New(&dbModel{}).Where("id = ?", 482000699).
		Scan(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbData)

	var dbDataList []map[string]any
	err = New(&dbModel{}).Limit("?", 3).
		Scan(&dbDataList).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbDataList)
}

func TestDb_Create(t *testing.T) {
	dbInit()

	var dbData = dbModel{
		Id:          1,
		Ts:          1,
		Uid:         "adsfasdfasdfasdfasdf",
		Ip:          "asdfas",
		HostIp:      "asdfasf",
		TestTime:    time.Now(),
		TestTimeMil: time.Now().AddDate(0, 0, 3),
		CreateId:    1,
		CreateTs:    1,
		CreateTime:  time.Now().AddDate(0, 0, 3),
		UpdateId:    2,
		UpdateTs:    2,
		UpdateTime:  time.Now().AddDate(0, 0, 3),
	}
	var err = New(&dbData).
		Create(&dbData).Error
	if err != nil {
		fmt.Println(err)
	}

	var dbDataPtr = &dbModel{
		Id:          1,
		Ts:          1,
		Uid:         "adsfasdfasdfasdfasdf",
		Ip:          "asdfas",
		HostIp:      "asdfasf",
		TestTime:    time.Now(),
		TestTimeMil: time.Now().AddDate(0, 0, 3),
		CreateId:    1,
		CreateTs:    1,
		CreateTime:  time.Now().AddDate(0, 0, 3),
		UpdateId:    2,
		UpdateTs:    2,
		UpdateTime:  time.Now().AddDate(0, 0, 3),
	}
	err = New(&dbData).
		Create(&dbDataPtr).Error
	if err != nil {
		fmt.Println(err)
	}
}

func TestDb_CreatePtr(t *testing.T) {
	dbInit()
	var dbData = dbModelPtr{
		Id: ToPtr(int64(0)),
		//Ts:          ToPtr(int64(1)),
		//Uid:         ToPtr("adsfasdfasdfasdfasdf"),
		//Ip:          ToPtr("asdfas"),
		HostIp:      ToPtr("asdfasf"),
		TestTime:    ToPtr(time.Now()),
		TestTimeMil: ToPtr(time.Now().AddDate(0, 0, 3)),
		CreateId:    ToPtr(int64(1)),
		CreateTs:    ToPtr(int64(1)),
		CreateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdateId:    ToPtr(int64(2)),
		UpdateTs:    ToPtr(int64(2)),
		UpdateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	var err = New(&dbData).
		Create(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_CreateMap(t *testing.T) {
	dbInit()
	var dbDataMap = map[string]any{
		"id":           0,
		"ts":           1,
		"uid":          "adsfasdfasdfasdfasdf",
		"ip":           "asdfas",
		"host_ip":      "asdfasf",
		"ip_long":      time.Now().Unix(),
		"peer_ip_long": time.Now().AddDate(0, 0, 3).UnixMilli(),
		"create_id":    1,
		"create_ts":    1,
		"create_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
		"update_id":    2,
		"update_ts":    2,
		"update_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
	}
	var err = New(&dbModel{}).
		Create(&dbDataMap).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_CreateInBatches(t *testing.T) {
	dbInit()
	var dbData = []*dbModel{
		&dbModel{
			Id:          0,
			Ts:          1,
			Uid:         "adsfasdfasdfasdfasdf",
			Ip:          "asdfas",
			HostIp:      "asdfasf",
			TestTime:    time.Now(),
			TestTimeMil: time.Now().AddDate(0, 0, 3),
			CreateId:    1,
			CreateTs:    1,
			CreateTime:  time.Now().AddDate(0, 0, 3),
			UpdateId:    2,
			UpdateTs:    2,
			UpdateTime:  time.Now().AddDate(0, 0, 3),
		},
		&dbModel{
			Id:          0,
			Ts:          1,
			Uid:         "adsfasdfasdfasdfasdf",
			Ip:          "asdfas",
			HostIp:      "asdfasf",
			TestTime:    time.Now(),
			TestTimeMil: time.Now().AddDate(0, 0, 3),
			CreateId:    1,
			CreateTs:    1,
			CreateTime:  time.Now().AddDate(0, 0, 3),
			UpdateId:    2,
			UpdateTs:    2,
			UpdateTime:  time.Now().AddDate(0, 0, 3),
		},
	}
	var err = New(&dbModel{}).
		CreateInBatches(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_CreateInBatchesPtr(t *testing.T) {
	dbInit()
	var dbData = []*dbModelPtr{
		&dbModelPtr{
			Id: ToPtr(int64(0)),
			//Ts:          ToPtr(int64(1)),
			//Uid:         ToPtr("adsfasdfasdfasdfasdf"),
			//Ip:          ToPtr("asdfas"),
			HostIp:      ToPtr("asdfasf"),
			TestTime:    ToPtr(time.Now()),
			TestTimeMil: ToPtr(time.Now().AddDate(0, 0, 3)),
			CreateId:    ToPtr(int64(1)),
			CreateTs:    ToPtr(int64(1)),
			CreateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
			UpdateId:    ToPtr(int64(2)),
			UpdateTs:    ToPtr(int64(2)),
			UpdateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		&dbModelPtr{
			Id: ToPtr(int64(0)),
			//Ts:          ToPtr(int64(1)),
			//Uid:         ToPtr("adsfasdfasdfasdfasdf"),
			//Ip:          ToPtr("asdfas"),
			HostIp:      ToPtr("asdfasf"),
			TestTime:    ToPtr(time.Now()),
			TestTimeMil: ToPtr(time.Now().AddDate(0, 0, 3)),
			CreateId:    ToPtr(int64(1)),
			CreateTs:    ToPtr(int64(1)),
			CreateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
			UpdateId:    ToPtr(int64(2)),
			UpdateTs:    ToPtr(int64(2)),
			UpdateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
		},
	}
	var err = New(&dbModel{}).
		CreateInBatches(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_CreateInBatchesMap(t *testing.T) {
	dbInit()
	var dbData = []map[string]any{
		{
			"id":           0,
			"ts":           1,
			"uid":          "adsfasdfasdfasdfasdf",
			"ip":           "asdfas",
			"host_ip":      "asdfasf",
			"ip_long":      time.Now().Unix(),
			"peer_ip_long": time.Now().AddDate(0, 0, 3).UnixMilli(),
			"create_id":    1,
			"create_ts":    1,
			"create_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
			"update_id":    2,
			"update_ts":    2,
			"update_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
		},
		{
			"id":           0,
			"ts":           1,
			"uid":          "adsfasdfasdfasdfasdf",
			"ip":           "asdfas",
			"host_ip":      "asdfasf",
			"ip_long":      time.Now().Unix(),
			"peer_ip_long": time.Now().AddDate(0, 0, 3).UnixMilli(),
			"create_id":    1,
			"create_ts":    1,
			"create_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
			"update_id":    2,
			"update_ts":    2,
			"update_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
		},
	}
	var err = New(&dbModel{}).
		CreateInBatches(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_Update(t *testing.T) {
	dbInit()
	var dbData = dbModel{
		Id:          0,
		Ts:          1,
		Uid:         "adsfasdfasdfasdfasdf",
		Ip:          "asdfas",
		HostIp:      "asdfasf",
		TestTime:    time.Now(),
		TestTimeMil: time.Now().AddDate(0, 0, 3),
		CreateId:    1,
		CreateTs:    1,
		CreateTime:  time.Now().AddDate(0, 0, 3),
		UpdateId:    2,
		UpdateTs:    2,
		UpdateTime:  time.Now().AddDate(0, 0, 3),
	}
	var err = New(&dbData).
		Where("id = ?", 1583993338).
		Updates(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_UpdatePtr(t *testing.T) {
	dbInit()
	var dbData = dbModelPtr{
		Id: ToPtr(int64(0)),
		//Ts:          ToPtr(int64(1)),
		//Uid:         ToPtr("adsfasdfasdfasdfasdf"),
		//Ip:          ToPtr("asdfas"),
		HostIp:      ToPtr("asdfasf"),
		TestTime:    ToPtr(time.Now()),
		TestTimeMil: ToPtr(time.Now().AddDate(0, 0, 3)),
		CreateId:    ToPtr(int64(1)),
		CreateTs:    ToPtr(int64(1)),
		CreateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdateId:    ToPtr(int64(2)),
		UpdateTs:    ToPtr(int64(2)),
		UpdateTime:  ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	var err = New(&dbData).
		Where("id = ?", 1583993338).
		Updates(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_UpdateMap(t *testing.T) {
	dbInit()
	var dbDataMap = map[string]any{
		"id":           0,
		"ts":           1,
		"uid":          "adsfasdfasdfasdfasdf",
		"ip":           "asdfas",
		"host_ip":      "asdfasf",
		"ip_long":      time.Now().Unix(),
		"peer_ip_long": time.Now().AddDate(0, 0, 3).UnixMilli(),
		"create_id":    1,
		"create_ts":    1,
		"create_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
		"update_id":    2,
		"update_ts":    2,
		"update_time":  time.Now().AddDate(0, 0, 3).Format(time.DateTime),
	}
	var err = New(&dbModel{}).
		Where("id = ?", 1583993338).
		Updates(&dbDataMap).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_UpdateRaw(t *testing.T) {
	dbInit()
	var dbDataMap = map[string]RawBody{
		"ts":        "1",
		"uid":       "adsfasdfasdfasdfasdf00",
		"ip":        "asdfas00",
		"host_ip":   "asdfasf00",
		"create_id": "1",
		"create_ts": "1",
		"update_id": "2",
		"update_ts": "2",
	}
	var err = New(&dbModel{}).
		Where("id = ?", 1583993338).
		Updates(&dbDataMap).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_CreateToSelect(t *testing.T) {
	dbInit()
	var err = New(&hhitAsset{}).
		Field("id,ts").
		Create(New(&hhitAsset{}).Field("id,ts").Where("id = ?", 4002258).Limit("1")).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestDb_ReadOnlyMap(t *testing.T) {
	// 1. 创建原始 map
	original := map[string]any{
		"name": "张三",
		"age":  18,
		"info": map[string]any{
			"hobby": "编程",
		},
	}

	// 2. 通过构造函数创建只读 map
	readOnly := NewReadOnlyMap(original)

	// 3. 读取操作（允许）
	if name, ok := readOnly.Get("name"); ok {
		fmt.Println("name:", name) // 输出：name: 张三
	}
	fmt.Println("所有键：", readOnly.Keys()) // 输出：所有键： [name age info]
	fmt.Println("长度：", readOnly.Len())   // 输出：长度： 3

	// 4. 遍历操作
	readOnly.Range(func(key string, val any) bool {
		fmt.Printf("遍历：%s = %v\n", key, val)
		return true // 继续遍历
	})
}

type Model struct {
	BaseMode
	Uuid string `json:"uuid"`
}

func (*Model) TableName() string {
	return "hhit_port"
}

// 流式查询测试
func TestDb_Stream(t *testing.T) {
	dbInit()
	var i = 0
	var err = New(Model{}).Where("id > 0").Stream(StreamCallback(func(t Model) (err error, next bool) {
		fmt.Println(t)
		i++
		return nil, true
	})).Error
	fmt.Println(err, i)
}

// 事务测试
func TestDb_Tx(t *testing.T) {
	dbInit()
	var tx, _ = Begin()
	var m, err = New("hhit_port").Tx(tx).Map()
	fmt.Println(m, err)
	m, err = New("hhit_asset").Tx(tx).Map()
	fmt.Println(m, err)
	err = New("hhit_asset").Tx(tx).Where("id <= 0").Limits(100).Delete().Error
	fmt.Println(err)
	err = New("hhit_asset").Tx(tx).Where("id <= 0").Updates(map[string]any{"id": 10}).Error
	fmt.Println(err)
	err = tx.Commit()
	fmt.Println(err)
}

// 事务测试
func TestDb_SelectAndCount(t *testing.T) {
	dbInit()
	var mdList []*Model
	var total int64
	var err = New(Model{}).LogLevel(DebugLogLevel).Where("id > 200").Page(1, 10).
		Group("port").
		SelectAndCount(&mdList, &total).Error
	fmt.Println(mdList)
	fmt.Println(total)
	fmt.Println(err)

	err = New(Model{}).LogLevel(DebugLogLevel).Where("id > 200").Page(1, 10).
		//Group("port").
		ScanAndCount(&mdList, &total).Error
	fmt.Println(mdList)
	fmt.Println(total)
	fmt.Println(err)

}
