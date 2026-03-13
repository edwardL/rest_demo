package gdbtmp

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type hhitAsset struct {
	BaseMode
	Uid         string `json:"uid"`          // uid
	BusinessIds []int  `json:"business_ids"` // 业务系统ID
	AreaIds     []int  `json:"area_ids"`     // 网络区域ID
	OrgIds      string `json:"org_ids"`      // 所属部门ID
}

// TableName 表名称
func (*hhitAsset) TableName() string {
	return "hhit_asset"
}

func gdbInit() {
	// 自定义数据类型  数据库 1,2,3 字符串转换为 结构体 []int
	var a = func(field reflect.StructField, fieldValue reflect.Value, itemMap map[string]any, tag GdbTag) (next bool) {
		// 判断 field 是否为 []int
		if field.Type == reflect.TypeOf([]int{}) {
			switch tag.Type {
			case "enum": // 根据不同类型处理成不同字符串 1,2,3
			case "json": // json [1,2,3]
			default:
				if intArr, ok := fieldValue.Interface().([]int); ok {
					var strArr []string = make([]string, len(intArr))
					for i, v := range intArr {
						strArr[i] = fmt.Sprintf("%d", v)
					}
					itemMap[tag.Name] = strings.Join(strArr, ",")
					return false // 不执行包内的默认处理
				}
				return false
			}
		}
		// 不是自定义类型 执行包内数据处理
		return true
	}
	var b = func(f reflect.Value, val any) (next bool, err error) {
		// 判断 field 是否为 []int
		if f.Type() == reflect.TypeOf([]int{}) {
			// 处理自定义类型
			var valStr = val.(string)
			arrVal := strings.Split(valStr, ",")
			var arrInt []int
			for _, v := range arrVal {
				valInt, _ := ToInt(v)
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
	Init(WithLogLevel(DebugLogLevel), WithMapAssignment(a), WithStructAssignment(b))
}

func TestDb_TypeScan(t *testing.T) {
	dbInit()
	gdbInit()
	var dbData hhitAsset
	var err = New(&dbData).Where("id = ?", 4002258).
		Scan(&dbData).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbData)

	var dbDataArr []*hhitAsset
	err = New(&dbData).Where("id IN (?)", []int{4002258, 4002272}).
		Scan(&dbDataArr).Error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(dbDataArr)
}

func TestDb_TypeUpdate(t *testing.T) {
	dbInit()
	gdbInit()
	var dbData hhitAsset = hhitAsset{
		BaseMode: BaseMode{
			Id:         4002258,
			Ts:         1,
			CreateId:   1,
			CreateTs:   1,
			CreateTime: "2024-11-14 10:30:02",
			UpdateId:   1,
			UpdateTs:   1,
			UpdateTime: "2025-09-02 14:42:25",
		},
		Uid:         "0a0a1fe20000000000000000000000000a0a1fe20000000000000000000000000000000000000000000000000000000000000000e01b4d56efd7655854c8e6ac231d7103000000000000000000000000",
		BusinessIds: []int{1, 7000001, 2000002, 3000009, 5000003},
		AreaIds:     []int{1, 10000004, 9000004},
		OrgIds:      "9,1",
	}
	var res, err = NewSql(&dbData).Where("id = ?", 4002258).
		Updates(&dbData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(res.CompSql())
}

func TestDb_TypeCreate(t *testing.T) {
	dbInit()
	gdbInit()
	var dbData hhitAsset = hhitAsset{
		BaseMode: BaseMode{
			Id:         4002258,
			Ts:         1,
			CreateId:   1,
			CreateTs:   1,
			CreateTime: "2024-11-14 10:30:02",
			UpdateId:   1,
			UpdateTs:   1,
			UpdateTime: "2025-09-02 14:42:25",
		},
		Uid:         "0a0a1fe20000000000000000000000000a0a1fe20000000000000000000000000000000000000000000000000000000000000000e01b4d56efd7655854c8e6ac231d7103000000000000000000000000",
		BusinessIds: []int{1, 7000001, 2000002, 3000009, 5000003},
		AreaIds:     []int{1, 10000004, 9000004},
		OrgIds:      "9,1",
	}
	var res, err = NewSql(&dbData).Where("id = ?", 4002258).
		Create(&dbData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(res.CompSql())
}

func TestDb_TypeCreateInBatches(t *testing.T) {
	dbInit()
	gdbInit()
	var dbData = []*hhitAsset{
		&hhitAsset{
			BaseMode: BaseMode{
				Id:         4002258,
				Ts:         1,
				CreateId:   1,
				CreateTs:   1,
				CreateTime: "2024-11-14 10:30:02",
				UpdateId:   1,
				UpdateTs:   1,
				UpdateTime: "2025-09-02 14:42:25",
			},
			Uid:         "0a0a1fe20000000000000000000000000a0a1fe20000000000000000000000000000000000000000000000000000000000000000e01b4d56efd7655854c8e6ac231d7103000000000000000000000000",
			BusinessIds: []int{1, 7000001, 2000002, 3000009, 5000003},
			AreaIds:     []int{1, 10000004, 9000004},
			OrgIds:      "9,1",
		},
		&hhitAsset{
			BaseMode: BaseMode{
				Id:         4002258,
				Ts:         1,
				CreateId:   1,
				CreateTs:   1,
				CreateTime: "2024-11-14 10:30:02",
				UpdateId:   1,
				UpdateTs:   1,
				UpdateTime: "2025-09-02 14:42:25",
			},
			Uid:         "0a0a1fe20000000000000000000000000a0a1fe20000000000000000000000000000000000000000000000000000000000000000e01b4d56efd7655854c8e6ac231d7103000000000000000000000000",
			BusinessIds: []int{1, 7000001, 2000002, 3000009, 5000003},
			AreaIds:     []int{1, 10000004, 9000004},
			OrgIds:      "9,1",
		},
	}
	var res, err = NewSql(&hhitAsset{}).Where("id = ?", 4002258).
		CreateInBatches(&dbData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Dump(res.CompSql())
}
