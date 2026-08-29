package mysql

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

const (
	DefaultMaxIdleConns = 10
	DefaultMaxOpenConns = 50
	DefaultMaxLifeTime  = 30 * time.Minute
)

// Config MySQL 连接配置（三方可直接构造，不依赖 gohera）。
type Config struct {
	MaxLifeTime  time.Duration `mapstructure:"max_life_time"`  // 连接可被重用的最长时间
	MaxOpenConns int           `mapstructure:"max_open_conns"` // 最大打开连接数
	MaxIdleConns int           `mapstructure:"max_idle_conns"` // 空闲池最大连接数
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Database     string        `mapstructure:"database"`
	Env          string        // DEV/TEST 时默认开 ShowSQL（可被 ShowSQL 覆盖）
	ShowSQL      *bool         // 显式控制 SQL 日志；nil 时回退 Env
	Charset      string        `mapstructure:"charset"` // 默认 utf8
}

// ConnectPool 兼容旧 API，内部仅持有配置。
type ConnectPool struct {
	config *Config
}

// DB 数据库连接，嵌入 xorm.Engine。
type DB struct {
	*xorm.Engine
	name string // 缓存键（Database）
}

var (
	dbMap = make(map[string]*xorm.Engine)
	dbMu  sync.RWMutex
)

// New 创建或复用 MySQL 连接。相同 Database 名全局复用。
func New(cfg *Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mysql: config is nil")
	}
	if err := cfg.validateAndNormalize(); err != nil {
		return nil, err
	}

	dbMu.RLock()
	if obj, ok := dbMap[cfg.Database]; ok {
		dbMu.RUnlock()
		return &DB{Engine: obj, name: cfg.Database}, nil
	}
	dbMu.RUnlock()

	dbMu.Lock()
	defer dbMu.Unlock()

	if obj, ok := dbMap[cfg.Database]; ok {
		return &DB{Engine: obj, name: cfg.Database}, nil
	}

	charset := cfg.Charset
	if charset == "" {
		charset = "utf8"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, charset)

	obj, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = obj.DB().Ping(); err != nil {
		_ = obj.Close()
		return nil, err
	}

	obj.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	obj.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	obj.DB().SetConnMaxLifetime(cfg.MaxLifeTime)
	obj.SetMapper(names.GonicMapper{})

	showSQL := cfg.resolveShowSQL()
	obj.ShowSQL(showSQL)
	if strings.EqualFold(cfg.Env, "DEV") || strings.EqualFold(cfg.Env, "TEST") {
		// 开发/测试缩短连接寿命，便于热配
		obj.DB().SetConnMaxLifetime(1 * time.Minute)
	}

	dbMap[cfg.Database] = obj
	return &DB{Engine: obj, name: cfg.Database}, nil
}

func (c *Config) validateAndNormalize() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("mysql: host is required")
	}
	if strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("mysql: database is required")
	}
	if strings.TrimSpace(c.User) == "" {
		return fmt.Errorf("mysql: user is required")
	}
	if c.Port <= 0 {
		c.Port = 3306
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = DefaultMaxIdleConns
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = DefaultMaxOpenConns
	}
	if c.MaxLifeTime <= 0 {
		c.MaxLifeTime = DefaultMaxLifeTime
	}
	return nil
}

func (c *Config) resolveShowSQL() bool {
	if c.ShowSQL != nil {
		return *c.ShowSQL
	}
	env := strings.ToUpper(c.Env)
	return env == "DEV" || env == "TEST"
}

// Bool 便于填写 Config.ShowSQL。
func Bool(v bool) *bool { return &v }

// InitOnce 兼容旧入口，等价于持有配置后调用 Connect→New。
func InitOnce(conf *Config) *ConnectPool {
	return &ConnectPool{config: conf}
}

// Connect 获取或创建数据库连接（兼容旧 API）。
func (o *ConnectPool) Connect() (*DB, error) {
	return New(o.config)
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

// CloseAll 关闭所有缓存中的 MySQL 连接。
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
