package repository

import (
	"context"
	"errors"
	"rest_demo/internal/model"

	"gorm.io/gorm"
)

type SysUserRepository interface {
	// PageList(ctx context.Context)
	// Create(ctx context.Context, user *model.SysUser) error
	GetByEmailOrUsername(ctx context.Context, email string, username string) (*model.SysUser, error)
}

func NewSysUserRepository(r *Repository) SysUserRepository {
	return &sysUserRepository{
		Repository: r,
	}
}

type sysUserRepository struct {
	*Repository
}

func (r *sysUserRepository) GetByEmailOrUsername(ctx context.Context, email string, username string) (*model.SysUser, error) {
	var user model.SysUser
	var err error
	if email != "" && username != "" {
		err = r.DB(ctx).Where("email = ? or username", email, username).First(&user).Error
	} else {
		if email != "" {
			err = r.DB(ctx).Where("email = ?", email).First(&user).Error
		} else if username != "" {
			err = r.DB(ctx).Where("username = ?", username).First(&user).Error
		} else {
			return nil, nil
		}
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
