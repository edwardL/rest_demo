package example

import (
	"fmt"
	"reflect"
	"rest_demo/pkg/gdbtmp"
	"strings"
)

// 自定义模型字段类型
// 添加模型字段类型对[]int的支持 数据库储存的为 1,2,3

// 通过初始化设置自定义参数
func gdbInit() {
	// 自定义数据类型  数据库 1,2,3 字符串转换为 结构体 []int
	var withMapAssignment = func(field reflect.StructField, fieldValue reflect.Value, itemMap map[string]any, tag gdbtmp.GdbTag) (next bool) {
		// 判断 field 是否为 []int
		if field.Type == reflect.TypeOf([]int{}) {
			switch tag.Type { // 模型的tag gdb的type部分 => `gdbtmp:"create_time;type:2006-01-02 15:04:05"`
			case "enum": // 根据不同类型处理成不同字符串 1,2,3
			case "json": // json [1,2,3]
			default: // 处理成1,2,3
				if intArr, ok := fieldValue.Interface().([]int); ok {
					var strArr []string = make([]string, len(intArr))
					for i, v := range intArr {
						strArr[i] = fmt.Sprintf("%d", v)
					}
					itemMap[tag.Name] = strings.Join(strArr, ",")
				}
			}
			return false // 不执行包内的默认处理
		}
		// 不是自定义类型 执行包内数据处理
		return true
	}
	var withStructAssignment = func(f reflect.Value, val any) (next bool, err error) {
		// 判断 field 是否为 []int
		if f.Type() == reflect.TypeOf([]int{}) {
			// 处理自定义类型
			var valStr = val.(string)
			arrVal := strings.Split(valStr, ",")
			var arrInt []int
			for _, v := range arrVal {
				valInt, _ := gdbtmp.ToInt(v)
				arrInt = append(arrInt, valInt)
			}
			// 设置值
			f.Set(reflect.ValueOf(arrInt))
			// 不再执行包内数据处理
			return false, nil
		}
		// 不是自定义类型 执行包内数据处理
		return true, nil
	}
	// 初始化注入自定义函数
	gdbtmp.Init(gdbtmp.WithMapAssignment(withMapAssignment), gdbtmp.WithStructAssignment(withStructAssignment))
}
