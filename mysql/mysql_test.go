package mysql

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}
	if err := cfg.validateAndNormalize(); err == nil {
		t.Fatal("expected error")
	}
	cfg = &Config{Host: "127.0.0.1", Database: "db", User: "u"}
	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3306 {
		t.Fatalf("port=%d", cfg.Port)
	}
	if cfg.MaxIdleConns != DefaultMaxIdleConns {
		t.Fatalf("idle=%d", cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns != DefaultMaxOpenConns {
		t.Fatalf("open=%d", cfg.MaxOpenConns)
	}
	if cfg.MaxLifeTime != DefaultMaxLifeTime {
		t.Fatalf("life=%v", cfg.MaxLifeTime)
	}
}

func TestResolveShowSQL(t *testing.T) {
	cfg := &Config{Env: "dev"}
	if !cfg.resolveShowSQL() {
		t.Fatal("dev should show sql")
	}
	cfg = &Config{Env: "prod", ShowSQL: Bool(true)}
	if !cfg.resolveShowSQL() {
		t.Fatal("explicit true")
	}
	cfg = &Config{Env: "dev", ShowSQL: Bool(false)}
	if cfg.resolveShowSQL() {
		t.Fatal("explicit false")
	}
}

func TestInitOnceCompat(t *testing.T) {
	p := InitOnce(&Config{Host: "h", Database: "d", User: "u", MaxLifeTime: time.Second})
	if p == nil || p.config == nil {
		t.Fatal("nil pool")
	}
}

func TestNewNilConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected error")
	}
}
