package easyorm

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestNewEngine_UnsupportedDialect(t *testing.T) {
	_, err := NewEngine("not_exists_driver", "dsn")
	if err == nil {
		t.Fatal("expected unsupported dialect error, got nil")
	}
}

func TestNewEngineWithOptions_SQLite(t *testing.T) {
	engine, err := NewEngineWithOptions("sqlite3", ":memory:", EngineOptions{SkipPing: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine == nil {
		t.Fatal("engine should not be nil")
	}
	if engine.Dialect() == nil {
		t.Fatal("engine dialect should not be nil")
	}

	if engine.NewSession() == nil {
		t.Fatal("session should not be nil")
	}

	engine.Close()
}
