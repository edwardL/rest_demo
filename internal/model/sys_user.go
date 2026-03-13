package model

import (
	"rest_demo/internal/constant"
	"rest_demo/pkg/db/field"

	"gorm.io/gorm"
)

type SysUser struct {
	UID           int64           `json:"uid" gorm:"column:uid;primaryKey;autoIncrement;comment:主键编码"`  //UID
	Username      string          `json:"username" gorm:"column:username;size:100;notNull;comment:用户名"` //用户名
	Password      []byte          `json:"-" gorm:"column:password;size:255;notNull;comment:密码"`
	NickName      string          `json:"nick_name" gorm:"column:nick_name;size:20;notNull;default:'';comment:昵称"` //昵称
	RememberToken string          `json:"-" gorm:"column:remember_token;size:500;notNull;default:'';comment:记住登陆token"`
	Internal      uint8           `json:"internal" gorm:"column:internal;notNull;default:1;comment:1内部用户"` //1内部用户
	AppUid        uint64          `json:"app_uid" gorm:"column:app_uid;notNull;default:1;comment:appuid"`  //外部用户 appuid
	Phone         string          `json:"phone" gorm:"column:phone;size:20;notNull;default:'';comment:手机"` //手机
	RoleId        int64           `json:"role_id" gorm:"column:role_id;notNull;default:0;comment:角色id"`
	Avatar        string          `json:"avatar" gorm:"column:avatar;size:255;notNull;default:'';comment:头像"` //头像
	Sex           int             `json:"sex" gorm:"column:sex;notNull;default:1;;comment:性别"`                //性别，1男2女
	Email         string          `json:"email" gorm:"column:email;size:100;notNull;default:'';comment:邮箱"`   //油箱
	DeptId        int64           `json:"dept_id" gorm:"column:dept_id;notNull;default:0;comment:部门id"`
	IsDefaultPwd  int64           `json:"is_default_pwd" gorm:"column:is_default_pwd;notNull;default:1;comment:是否需要重置密码1是2否"` //是否需要重置密码1是2否
	Remark        string          `json:"remark" gorm:"column:remark;notNull;default:'';comment:备注"`                          //备注
	Status        constant.Status `json:"status" gorm:"column:status;notNull;default:1;comment:状态"`                           //状态 1启用2禁用
	CreateBy      int64           `json:"create_by" gorm:"column:create_by;notNull;default:0;comment:创建者"`                    //创建者
	UpdateBy      int64           `json:"update_by" gorm:"column:update_by;notNull;default:0;comment:更新者"`                    //最后更新者
	CreatedAt     field.LocalTime `json:"created_at" gorm:"column:created_at;notNull"`
	UpdatedAt     field.LocalTime `json:"updated_at" gorm:"column:updated_at;notNull"`
	DeletedAt     gorm.DeletedAt  `json:"-" gorm:"column:deleted_at;index;"`
}

func (s *SysUser) TableName() string {
	return "sys_user"
}
