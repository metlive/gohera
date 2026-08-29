package sqlite

import (
	"context"
	"fmt"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

const (
	DefaultMaxIdleConns = 5
	DefaultMaxOpenConns = 10
)

// Config SQLite 配置（三方可直接构造，不依赖 gohera）。
type Config struct {
	FilePath     string `mapstructure:"file_path"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	ShowSQL      bool   `mapstructure:"show_sql"`
}

// DB 数据库连接，嵌入 xorm.Engine。
type DB struct {
	*xorm.Engine
	name string // 缓存键（FilePath）
}

var (
	dbMap = make(map[string]*xorm.Engine)
	dbMu  sync.RWMutex
)

// New 创建或复用 SQLite 连接。相同 FilePath 全局复用。
func New(cfg *Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sqlite: config is nil")
	}
	if err := cfg.validateAndNormalize(); err != nil {
		return nil, err
	}

	dbMu.RLock()
	if obj, ok := dbMap[cfg.FilePath]; ok {
		dbMu.RUnlock()
		return &DB{Engine: obj, name: cfg.FilePath}, nil
	}
	dbMu.RUnlock()

	dbMu.Lock()
	defer dbMu.Unlock()

	if obj, ok := dbMap[cfg.FilePath]; ok {
		return &DB{Engine: obj, name: cfg.FilePath}, nil
	}

	obj, err := xorm.NewEngine("sqlite3", cfg.FilePath)
	if err != nil {
		return nil, err
	}
	if err = obj.DB().Ping(); err != nil {
		_ = obj.Close()
		return nil, err
	}

	obj.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	obj.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	obj.SetMapper(names.GonicMapper{})
	obj.ShowSQL(cfg.ShowSQL)

	dbMap[cfg.FilePath] = obj
	return &DB{Engine: obj, name: cfg.FilePath}, nil
}

func (c *Config) validateAndNormalize() error {
	if strings.TrimSpace(c.FilePath) == "" {
		return fmt.Errorf("sqlite: file_path is required")
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = DefaultMaxIdleConns
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = DefaultMaxOpenConns
	}
	return nil
}

// Close 关闭本连接并从缓存移除。
func (db *DB) Close() error {
	if db == nil || db.Engine == nil {
		return nil
	}
	dbMu.Lock()
	defer dbMu.Unlock()
	if db.name != "" {
		if cur, ok := dbMap[db.name]; ok && cur == db.Engine {
			delete(dbMap, db.name)
		}
	}
	return db.Engine.Close()
}

// CloseAll 关闭所有缓存中的 SQLite 连接。
func CloseAll() error {
	dbMu.Lock()
	defer dbMu.Unlock()
	var firstErr error
	for name, eng := range dbMap {
		if err := eng.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(dbMap, name)
	}
	return firstErr
}

// Context 返回带有上下文的 Session。
func (db *DB) Context(ctx context.Context) *Session {
	return &Session{db.Engine.Context(ctx)}
}
