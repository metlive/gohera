package sqlite

import (
	"context"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

// Config SQLite 配置
type Config struct {
	FilePath     string `toml:"file_path"`      // 数据库文件路径
	MaxOpenConns int    `toml:"max_open_conns"` // 设置打开连接到数据库的最大数量
	MaxIdleConns int    `toml:"max_idle_conns"` // 设置空闲连接池中的最大连接数
	ShowSQL      bool   `toml:"show_sql"`       // 是否开启 SQL 日志
}

// DB 数据库连接，嵌入 xorm.Engine
type DB struct {
	*xorm.Engine
}

var (
	dbMap = make(map[string]*xorm.Engine)
	dbMu  sync.RWMutex
)

// New 创建或获取 SQLite 数据库连接
// 相同 filePath 会复用已有连接
func New(cfg *Config) (*DB, error) {
	dbMu.RLock()
	if obj, ok := dbMap[cfg.FilePath]; ok {
		dbMu.RUnlock()
		return &DB{obj}, nil
	}
	dbMu.RUnlock()

	dbMu.Lock()
	defer dbMu.Unlock()

	// Double check
	if obj, ok := dbMap[cfg.FilePath]; ok {
		return &DB{obj}, nil
	}

	obj, err := xorm.NewEngine("sqlite3", cfg.FilePath)
	if err != nil {
		return nil, err
	}

	if cfg.MaxIdleConns > 0 {
		obj.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		obj.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	}
	obj.SetMapper(names.GonicMapper{})
	obj.ShowSQL(cfg.ShowSQL)

	dbMap[cfg.FilePath] = obj
	return &DB{obj}, nil
}

// Context 返回带有上下文的 Session
func (db *DB) Context(ctx context.Context) *Session {
	return &Session{db.Engine.Context(ctx)}
}
