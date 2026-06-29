package conv

import (
	"reflect"
	"sync"
)

type fieldInfo struct {
	index     int
	name      string
	structTag StructTag
	fieldType reflect.Type
	anonymous bool
	isPtr     bool
	kind      reflect.Kind
}

// fieldCache 缓存结构体类型到字段信息的映射，使用LRU 策略限制大小
type fieldCacheType struct {
	mu      sync.RWMutex
	items   map[reflect.Type][]fieldInfo
	order   []reflect.Type // 记录访问顺序
	maxSize int            // 最大缓存大小
}

func newFieldCache(maxsize int) *fieldCacheType {
	return &fieldCacheType{
		items:   make(map[reflect.Type][]fieldInfo),
		order:   make([]reflect.Type, 0),
		maxSize: maxsize,
	}
}

var fieldCache = newFieldCache(10000) // 限制缓存最多10000种类型 200万个字段 大约 200m
// getStructFieldInfo 获取结构体字段信息，带缓存
func getStructFieldInfo(t reflect.Type) []fieldInfo {
	// 先尝试从缓存中获取
	fieldCache.mu.RLock()
	if info, exists := fieldCache.items[t]; exists {
		fieldCache.mu.RUnlock()
		return info
	}
	fieldCache.mu.RUnlock()

	fieldCache.mu.Lock()
	defer fieldCache.mu.Unlock()
	// 再次检查，万一命中了勒
	if info, exists := fieldCache.items[t]; exists {
		return info
	}

	var fieldInfoList = make([]fieldInfo, 0, t.NumField())
	var field reflect.StructField
	var tag StructTag
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		tag = getStructTag(field)
		if tag.Name == "-" {
			continue
		}
		fieldInfoList = append(fieldInfoList, fieldInfo{
			index:     i,
			name:      field.Name,
			structTag: tag,
			fieldType: field.Type,
			anonymous: field.Anonymous,
			isPtr:     field.Type.Kind() == reflect.Ptr,
			kind:      field.Type.Kind(),
		})
	}

	// 检查是否需要清理缓存
	if len(fieldCache.items) >= fieldCache.maxSize {
		// 移除最早添加的项
		if len(fieldCache.order) > 0 {
			delete(fieldCache.items, fieldCache.order[0])
			fieldCache.order = fieldCache.order[1:]
		}
	}
	// 添加新项
	fieldCache.items[t] = fieldInfoList
	fieldCache.order = append(fieldCache.order, t)

	return fieldInfoList
}

type StructTag struct {
	Name string
	Type string
}

func getStructTag(field reflect.StructField) StructTag {
	tag := field.Tag.Get("json")
	return StructTag{Name: tag, Type: field.Type.String()}
}
