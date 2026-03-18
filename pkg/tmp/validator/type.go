package validator

type TagOptions struct {
	Valid            string `json:"valid_tag"`           // 验证规则标签
	ValidCode        string `json:"valid_code"`          // 验证规则错误码
	ValidMsg         string `json:"valid_msg"`           // 验证规则错误信息
	ValidSep         byte   `json:"valid_sep"`           // 验证规则分隔符
	ValidOptSep      byte   `json:"valid_opt_sep"`       // 多规则选项分隔符
	GroupOptSepLeft  byte   `json:"group_opt_sep_left"`  // 依赖规则选项组分隔符
	GroupOptSepRight byte   `json:"group_opt_sep_right"` // 依赖规则选项组分隔符
	GroupOptElseSep  byte   `json:"group_opt_else_sep"`  // 验证规则选项组 规则分隔符
}

var defTagOptions = TagOptions{
	Valid:            "v",
	ValidCode:        "vCode",
	ValidMsg:         "vMsg",
	ValidSep:         '|',
	ValidOptSep:      ':',
	GroupOptSepLeft:  '[',
	GroupOptSepRight: ']',
	GroupOptElseSep:  ',',
}
