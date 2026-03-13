package migration

import (
	"rest_demo/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Config struct {
}

type Migration struct {
	config *Config
	db     *gorm.DB
	log    *zap.Logger
}

func NewMigration(conf *Config, log *zap.Logger, db *gorm.DB) *Migration {
	return &Migration{
		config: conf,
		log:    log,
		db:     db,
	}
}

func (m *Migration) InitTables() error {
	return m.db.AutoMigrate(
		&model.SysUser{},
	)
}
