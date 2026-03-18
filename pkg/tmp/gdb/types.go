package gdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ModelFace 模型结构体接口
type ModelFace interface {
	TableName() string
}

// ModelAliasFace 模型结构体接口
type ModelAliasFace interface {
	TableAlias() string
}

// modelFaceType ModelFace接口的反射类型
var modelFaceType = reflect.TypeOf((*ModelFace)(nil)).Elem()

// modelAliasFaceType ModelFace接口的反射类型
var modelAliasFaceType = reflect.TypeOf((*ModelAliasFace)(nil)).Elem()

// BaseModeTime 模型基础字段 带时间类型
type BaseModeTime struct {
	Id         int       `json:"id"`                                                     // ID
	Ts         int       `json:"ts"`                                                     // TS
	CreateId   int       `json:"create_id"`                                              // 创建用户ID
	CreateTs   int       `json:"create_ts"`                                              // 创建用户TS
	CreateTime time.Time `json:"create_time" gdb:"create_time;type:2006-01-02 15:04:05"` // 创建时间
	UpdateId   int       `json:"update_id"`                                              // 更新用户ID
	UpdateTs   int       `json:"update_ts"`                                              // 更新用户TS
	UpdateTime time.Time `json:"update_time" gdb:"update_time;type:2006-01-02 15:04:05"` // 更新时间
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

type GdbTag = conv.StructTag

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
	call(data map[string]any, cols []string, dbConvInitPtr bool) (next bool, err error)
	make() any
}

// 流式查询回调函数
type streamCallback[T any] func(t T) (next bool, err error)

// 获取当前类型实例
func (c streamCallback[T]) make() any {
	var t T
	return t
}

// 执行回调函数
func (c streamCallback[T]) call(data map[string]any, cols []string, dbConvInitPtr bool) (next bool, err error) {
	var res = c.make().(T)
	if _, ok := any(res).(map[string]any); ok {
		return c(any(data).(T))
	}
	var convErr = conv.MapToStruct(data, &res, dbConvInitPtr)
	next, err = c(res)
	return next, errors.Join(convErr, err)
}

// KeyValue 键值对结构体
type KeyValue[K comparable, T any] struct {
	Key   K
	Value T
}

// OrderedMap 定义有序map结构
type OrderedMap[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

// NewOrderedMap 创建新的有序map
func NewOrderedMap[K comparable, T any]() *OrderedMap[K, T] {
	return &OrderedMap[K, T]{
		keys:   make([]K, 0),
		values: make(map[K]T),
	}
}

// SetKey 设置键
func (om *OrderedMap[K, T]) SetKey(key []K) {
	om.keys = key
}

// SetValues 设置键
func (om *OrderedMap[K, T]) SetValues(values map[K]T) {
	om.values = values
}

// Set 设置键值对
func (om *OrderedMap[K, T]) Set(key K, value T) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value
}

// Get 获取指定key的值
func (om *OrderedMap[K, T]) Get(key K) (T, bool) {
	value, exists := om.values[key]
	return value, exists
}

// Delete 删除指定key
func (om *OrderedMap[K, T]) Delete(key K) {
	if _, exists := om.values[key]; exists {
		delete(om.values, key)
		for i, k := range om.keys {
			if k == key {
				om.keys = append(om.keys[:i], om.keys[i+1:]...)
				break
			}
		}
	}
}

// Keys 返回所有的key，保持插入顺序
func (om *OrderedMap[K, T]) Keys() []K {
	return append([]K{}, om.keys...)
}

// Values 返回所有的值，保持插入顺序
func (om *OrderedMap[K, T]) Values() []T {
	values := make([]T, len(om.keys))
	for i, key := range om.keys {
		values[i] = om.values[key]
	}
	return values
}

// All 实现range over func语法，支持 yield
func (om *OrderedMap[K, T]) All(yield func(K, T) bool) {
	for _, key := range om.keys {
		if !yield(key, om.values[key]) {
			return
		}
	}
}

// AllPairs 实现range over func语法，返回键值对结构体
func (om *OrderedMap[K, T]) AllPairs(yield func(value KeyValue[K, T]) bool) {
	for _, key := range om.keys {
		kv := KeyValue[K, T]{Key: key, Value: om.values[key]}
		if !yield(kv) {
			return
		}
	}
}

// MarshalJSON 实现json.Marshaler接口，提供自定义JSON序列化
func (om *OrderedMap[K, T]) MarshalJSON() ([]byte, error) {
	if om.keys == nil && om.values == nil {
		return []byte("{}"), nil
	}

	// 使用数组形式保证顺序
	var result = make([]byte, 0, 2)
	result = append(result, '{')
	for i, key := range om.keys {
		if i > 0 {
			result = append(result, ',')
		}
		// 添加键（需要转义）
		keyBytes, _ := json.Marshal(key)
		result = append(result, keyBytes...)
		result = append(result, ':')

		// 添加值
		valBytes, err := json.Marshal(om.values[key])
		if err != nil {
			return nil, err
		}
		result = append(result, valBytes...)
	}
	result = append(result, '}')
	return result, nil
}

// UnmarshalJSON 实现json.Unmarshaler接口，提供自定义JSON反序列化
func (om *OrderedMap[K, T]) UnmarshalJSON(data []byte) error {
	// 先重置当前对象
	om.keys = make([]K, 0)
	om.values = make(map[K]T)

	// 解析JSON到临时map
	var temp map[K]T
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// 按照顺序添加元素
	for key, value := range temp {
		om.Set(key, value)
	}

	return nil
}

// StreamCallback 构建一个流式查询入参
func StreamCallback[T any](f func(t T) (next bool, err error)) streamCallbackFace {
	return streamCallback[T](f)
}

type Page[T any] struct {
	CurrPage  int  // 当前页数
	PageNums  int  // 分页数量
	TotalNums int  // 总条数
	PageData  []*T // 分页数据
}

// NewPage 创建分页对象
func NewPage[T any](currPage int, pageNums int) Page[T] {
	return Page[T]{
		CurrPage:  currPage,
		PageNums:  pageNums,
		TotalNums: 0,
		PageData:  nil,
	}
}

// Wrapper 查询包装器
type Wrapper struct {
	*Db
}

// QueryWrapper 创建查询包装器
func QueryWrapper() *Wrapper {
	return &Wrapper{Db: &Db{gs: NewSql()}}
}

// GetSql sql构建
func (w *Wrapper) GetSql() string {
	var res, _ = w.gs.Select()
	if res != nil {
		return res.Sql.String()
	}
	return ""
}

// GetArgs sql构建
func (w *Wrapper) GetArgs() []any {
	var res, _ = w.gs.Select()
	if res != nil {
		return res.Args
	}
	return make([]any, 0)
}

// GetWhereSql 返回当前条件构造出的 WHERE SQL 片段。
func (w *Wrapper) GetWhereSql() string {
	var query, _ = w.Db.GenWhere()
	return query
}

// GetWhereArgs 返回当前条件构造出的 WHERE 参数列表。
func (w *Wrapper) GetWhereArgs() []any {
	var _, args = w.Db.GenWhere()
	return args
}

// JSONType 通用 JSON 字段类型。
type JSONType[T any] struct {
	Data T
}

// Value 将 Data 序列化为 JSON，供数据库写入（实现 driver.Valuer）。
func (j JSONType[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}

// Scan 将数据库值反序列化到 Data（实现 sql.Scanner）。
func (j *JSONType[T]) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	case T:
		j.Data = v
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) == 0 {
		j.Data = *(new(T))
		return nil
	}
	return json.Unmarshal(bytes, &j.Data)
}

// MarshalJSON 按 Data 的 JSON 结构输出。
func (j JSONType[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON 从 JSON 字节反序列化到 Data。
func (j *JSONType[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &j.Data)
}

// ENUMType 通用枚举数组类型（当前支持 int/int8/string）。
type ENUMType[T int | int8 | string] []T

// Value 将枚举数组序列化为逗号分隔字符串（实现 driver.Valuer）。
func (j ENUMType[T]) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "", nil
	}

	var strArr []string = make([]string, len(j))
	var idx int
	for idx = 0; idx < len(j); idx++ {
		strArr[idx] = fmt.Sprint(j[idx])
	}

	return strings.Join(strArr, ","), nil
}

// Scan 将数据库值反序列化为枚举数组（实现 sql.Scanner）。
func (j *ENUMType[T]) Scan(value interface{}) error {
	var val string
	switch value.(type) {
	case []byte:
		val = string(value.([]byte))
	case string:
		val = value.(string)
	default:
		return fmt.Errorf("failed to scan enum value: %v", value)
	}

	if len(strings.TrimSpace(val)) == 0 {
		*j = make([]T, 0)
		return nil
	}

	var strArr []string = strings.Split(val, ",")
	var result []T = make([]T, 0, len(strArr))
	var zero T
	var item string
	var parsed int
	var parsedInt int64
	var err error
	switch any(zero).(type) {
	case string:
		for _, item = range strArr {
			var enumVal any = strings.TrimSpace(item)
			result = append(result, enumVal.(T))
		}
	case int:
		for _, item = range strArr {
			parsed, err = strconv.Atoi(strings.TrimSpace(item))
			if err != nil {
				return fmt.Errorf("failed to parse int enum value %q: %w", item, err)
			}
			var enumVal any = parsed
			result = append(result, enumVal.(T))
		}
	case int8:
		for _, item = range strArr {
			parsedInt, err = strconv.ParseInt(strings.TrimSpace(item), 10, 8)
			if err != nil {
				return fmt.Errorf("failed to parse int8 enum value %q: %w", item, err)
			}

			var enumVal any = int8(parsedInt)
			result = append(result, enumVal.(T))
		}
	default:
		return fmt.Errorf("unsupported enum type")
	}

	*j = result
	return nil
}
