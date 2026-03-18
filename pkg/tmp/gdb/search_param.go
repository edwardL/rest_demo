package gdb

import (
	"fmt"
	hhtools "nwgit.gzhhit.com/BD/hhitcomm.git"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
)

// HhSearchParam 搜索参数处理器，函数式处理前端分页和搜索参数
type HhSearchParam types.SearchParams

// NewSearchParamsHandle 创建搜索参数处理
func NewSearchParamsHandle(sp *types.SearchParams) *HhSearchParam {
	var sph *HhSearchParam = (*HhSearchParam)(sp)
	return sph
}

// GetSearchParams 获取搜索参数
func (p *HhSearchParam) GetSearchParams() *types.SearchParams {
	return (*types.SearchParams)(p)
}

// GetSearchItem 获取搜索参数条件
func (p *HhSearchParam) GetSearchItem() []types.SearchItem {
	return p.SearchParams
}

// GetSearchValue 获取查询参数的值
func (p *HhSearchParam) GetSearchValue(fieldName string) (any, error) {
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			return p.SearchParams[k].FieldValue, nil
		}
	}
	return 0, fmt.Errorf("未发现查询参数的值 %s", fieldName)
}

// GetSearchIntValue 获取查询参数的值
func (p *HhSearchParam) GetSearchIntValue(fieldName string) (int, error) {
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			return conv.ToInt(p.SearchParams[k].FieldValue)
		}
	}
	return 0, fmt.Errorf("未发现查询参数的值 %s", fieldName)
}

// GetSearchStringValue 获取查询参数的值
func (p *HhSearchParam) GetSearchStringValue(fieldName string) (string, error) {
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			return conv.ToString(p.SearchParams[k].FieldValue), nil
		}
	}
	return "", fmt.Errorf("未发现查询参数的值 %s", fieldName)
}

// GetSearchOp 获取查询参数的操作符
func (p *HhSearchParam) GetSearchOp(fieldName string) (string, error) {
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			return p.SearchParams[k].FieldCmpOp, nil
		}
	}
	return "", fmt.Errorf("未发现查询参数的值 %s", fieldName)
}

// ReplaceFieldName 替换搜索字段名
func (p *HhSearchParam) ReplaceFieldName(oldFieldName string, newFieldName string) {
	if len(p.SearchParams) == 0 {
		return
	}
	for i := 0; i < len(p.SearchParams); i++ {
		if p.SearchParams[i].FieldName == oldFieldName {
			p.SearchParams[i].FieldName = newFieldName
		}
	}
}

// ReplaceFieldValue 替换搜索字段值
func (p *HhSearchParam) ReplaceFieldValue(fieldName string, f func(any) any) {
	if len(p.SearchParams) == 0 {
		return
	}
	var oldFieldValue, err = p.GetSearchValue(fieldName)
	if err != nil {
		return
	}
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			p.SearchParams[k].FieldValue = f(oldFieldValue)
		}
	}
}

// RemoveSearch 删除搜索字段
func (p *HhSearchParam) RemoveSearch(fieldNames ...string) {
	if len(p.SearchParams) == 0 {
		return
	}
	var searchParams []types.SearchItem
	for _, searchParam := range p.SearchParams {
		var contain bool
		for _, fieldName := range fieldNames {
			if fieldName == searchParam.FieldName {
				contain = true
				break
			}
		}
		if !contain {
			searchParams = append(searchParams, searchParam)
		}
	}
	p.SearchParams = searchParams
}

// KeepSearch 保留合法的搜索字段
func (p *HhSearchParam) KeepSearch(fieldNames ...string) {
	if len(p.SearchParams) == 0 {
		return
	}
	var searchParams []types.SearchItem
	for k := range p.SearchParams {
		for _, fieldName := range fieldNames {
			if fieldName == p.SearchParams[k].FieldName {
				searchParams = append(searchParams, p.SearchParams[k])
			}
		}
	}
	p.SearchParams = searchParams
}

// ReplaceSort 保留合法的排序
func (p *HhSearchParam) ReplaceSort(rs map[string]string) {
	if len(p.SortParams) == 0 {
		return
	}
	p.SortParams, _, _ = ValidateOrderParam(p.SortParams, rs)
}

// KeepSort 保留合法的排序
func (p *HhSearchParam) KeepSort(sorts ...string) {
	var sortsMap = map[string]string{}
	for _, sort := range sorts {
		sortsMap[sort] = sort
	}
	p.ReplaceSort(sortsMap)
}

// RequiredSearch 必须的参数
func (p *HhSearchParam) RequiredSearch(fieldName string, msg string) error {
	for k := range p.SearchParams {
		if p.SearchParams[k].FieldName == fieldName {
			return nil
		}
	}
	return fmt.Errorf("%s: %s", hhtools.CodeReqParamsError, msg)
}

// AddSearch 添加参数
func (p *HhSearchParam) AddSearch(fieldName string, fieldCmpOp string, fieldValue any) *HhSearchParam {
	p.SearchParams = append(p.SearchParams, types.SearchItem{
		FieldName:  fieldName,
		FieldCmpOp: fieldCmpOp,
		FieldValue: fieldValue,
	})
	return p
}
