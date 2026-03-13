package example

import (
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
func (dbModel) TableName() string {
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
func (dbModelPtr) TableName() string {
	return "`hhit_asset`.`hhit_agent_network`"
}

type hhitAsset struct {
	gdb.BaseMode
	Uid         string `json:"uid"`          // uid
	BusinessIds string `json:"business_ids"` // 业务系统ID
	AreaIds     string `json:"area_ids"`     // 网络区域ID
	OrgIds      string `json:"org_ids"`      // 所属部门ID
}
