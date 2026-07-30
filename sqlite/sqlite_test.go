package sqlite

import (
	"path/filepath"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}
	if err := cfg.validateAndNormalize(); err == nil {
		t.Fatal("expected error")
	}
	cfg = &Config{FilePath: "a.db"}
	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatal(err)
	}
	if cfg.MaxIdleConns != DefaultMaxIdleConns || cfg.MaxOpenConns != DefaultMaxOpenConns {
		t.Fatalf("defaults idle=%d open=%d", cfg.MaxIdleConns, cfg.MaxOpenConns)
	}
}

func TestNewCachedByFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	t.Cleanup(func() { _ = CloseAll() })

	db1, err := New(&Config{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	db2, err := New(&Config{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	if db1.Engine != db2.Engine {
		t.Fatal("expected same cached engine")
	}
	if err := db1.Close(); err != nil {
		t.Fatal(err)
	}
	db3, err := New(&Config{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	if db3.Engine == db1.Engine {
		t.Fatal("expected new engine after close")
	}
	_ = db3.Close()
}

func TestNewNilConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected error")
	}
}
