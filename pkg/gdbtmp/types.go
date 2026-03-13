package gdbtmp

import (
	"errors"
	"reflect"
	"time"
)

// ModelFace 模型结构体接口
type ModelFace interface {
	TableName() string
}

// modelFaceType ModelFace接口的反射类型
var modelFaceType = reflect.TypeOf((*ModelFace)(nil)).Elem()

// structFieldInfo 字段消息
type structFieldInfo struct {
	Name      string
	Anonymous bool
}

// BaseModeTime 模型基础字段 带时间类型
type BaseModeTime struct {
	Id         int       `json:"id"`                                                        // ID
	Ts         int       `json:"ts"`                                                        // TS
	CreateId   int       `json:"create_id"`                                                 // 创建用户ID
	CreateTs   int       `json:"create_ts"`                                                 // 创建用户TS
	CreateTime time.Time `json:"create_time" gdbtmp:"create_time;type:2006-01-02 15:04:05"` // 创建时间
	UpdateId   int       `json:"update_id"`                                                 // 更新用户ID
	UpdateTs   int       `json:"update_ts"`                                                 // 更新用户TS
	UpdateTime time.Time `json:"update_time" gdbtmp:"update_time;type:2006-01-02 15:04:05"` // 更新时间
}

// BaseMode 模型基础字段
type BaseMode struct {
	Id         int    `json:"id"`          // ID
	Ts         int    `json:"ts"`          // TS
	CreateId   int    `json:"create_id"`   // 创建用户ID
	CreateTs   int    `json:"create_ts"`   // 创建用户TS
	CreateTime string `json:"create_time"` // 创建时间
	UpdateId   int    `json:"update_id"`   // 更新用户ID
	UpdateTs   int    `json:"update_ts"`   // 更新用户TS
	UpdateTime string `json:"update_time"` // 更新时间
}

// ReadOnlyMap 只读的map
type ReadOnlyMap struct {
	data map[string]any
}

// NewReadOnlyMap 创建只读的map
func NewReadOnlyMap(m map[string]any) *ReadOnlyMap {
	if m == nil {
		m = make(map[string]any)
	}
	return &ReadOnlyMap{m}
}

// Get 获取键对应的值，返回值和是否存在
func (m *ReadOnlyMap) Get(key string) (any, bool) {
	var val, ok = m.data[key]
	return val, ok
}

// Keys 返回所有键的列表
func (m *ReadOnlyMap) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Len 返回 map 长度
func (m *ReadOnlyMap) Len() int {
	return len(m.data)
}

// Range 遍历键值对，通过回调函数处理（类似 sync.Map 的 Range）
// 若回调返回 false，停止遍历
func (m *ReadOnlyMap) Range(fn func(key string, val any) bool) {
	for k, v := range m.data {
		if !fn(k, v) {
			break
		}
	}
}

// 流式查询入参
type streamCallbackFace interface {
	call(data, cols []string) (err error, next bool)
	make() any
}

// 流式查询回调函数
type streamCallback[T any] func(t T) (err error, next bool)

// 获取当前类型实例
func (c streamCallback[T]) make() any {
	var t T
	return t
}

// 执行回调函数
func (c streamCallback[T]) call(data, cols []string) (err error, next bool) {
	var res = c.make().(T)
	if _, ok := any(res).(map[string]any); ok {
		var m = make(map[string]any, len(cols))
		for k, v := range data {
			m[cols[k]] = v
		}
		return c(any(m).(T))
	}
	var convErr = SliceToStruct([][]string{cols, data}, &res)
	err, next = c(res)
	return errors.Join(convErr, err), next
}

// StreamCallback 构建一个流式查询入参
func StreamCallback[T any](f func(t T) (err error, next bool)) streamCallbackFace {
	return streamCallback[T](f)
}

const (
	TagTypeKey           = "type"      // 模型标签类型标识
	TagTypeTimeUnixMilli = "UnixMilli" // 毫秒时间戳
	TagTypeTimeUnix      = "Unix"      // 秒时间戳
)

var (
	TimeType        = reflect.TypeOf(time.Time{})
	TimePtrType     = reflect.TypeOf(&time.Time{})
	notStructErrMsg = "value must be a struct or pointer to struct"
)
