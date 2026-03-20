package easyorm

import (
	"database/sql"
	"fmt"
	"rest_demo/pkg/easyorm/dialect"
	"rest_demo/pkg/easyorm/log"
	"rest_demo/pkg/easyorm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

type EngineOptions struct {
	SkipPing bool
}

func NewEngine(driver, source string) (e *Engine, err error) {
	return NewEngineWithOptions(driver, source, EngineOptions{})
}

func NewEngineWithOptions(driver, source string, opts EngineOptions) (e *Engine, err error) {
	d, ok := dialect.GetDialect(driver)
	if !ok {
		err = fmt.Errorf("easyorm: unsupported dialect %q", driver)
		log.Error(err)
		return
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	if !opts.SkipPing {
		if err = db.Ping(); err != nil {
			log.Error(err)
			return
		}
	}

	e = &Engine{db: db, dialect: d}
	log.Info("Connect database success")
	return
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Error(err)
	}
	log.Info("Close database success")
}

func (e *Engine) NewSession() *session.Session {
	return session.New(e.db)
}

func (e *Engine) Dialect() dialect.Dialect {
	return e.dialect
}
