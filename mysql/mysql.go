package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

type Config struct {
	MaxLifeTime  time.Duration `toml:"max_life_time"`  // 设置连接可以被重新使用的最大时间量
	MaxOpenConns int           `toml:"max_open_conns"` // 设置打开连接到数据库的最大数量
	MaxIdleConns int           `toml:"max_idle_conns"` // 设置空闲连接池中的最大连接数
	User         string        `toml:"user"`           //用户名
	Password     string        `toml:"password"`       //密码
	Host         string        `toml:"host"`           //数据库地址
	Port         int           `toml:"port"`           //端口
	Database     string        `toml:"database"`       //连接那个数据库
	Env          string
}

// 变量初始化
type ConnectPool struct {
	mutex  sync.Mutex
	config *Config
}
type DB struct {
	*xorm.EngineGroup
}

var (
	dbMap    = make(map[string]*xorm.EngineGroup)
	instance *ConnectPool
	once     sync.Once
)

func InitOnce(conf *Config) *ConnectPool {
	once.Do(func() {
		instance = &ConnectPool{
			config: conf,
		}
	})
	return instance
}

func (o *ConnectPool) Connect() (*DB, error) {
	if obj, ok := dbMap[o.config.Database]; ok {
		return &DB{obj}, nil
	} else {
		var dataSource []string
		hosts := strings.Split(o.config.Host, ",")
		for _, host := range hosts {
			dataSource = append(dataSource, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", o.config.User, o.config.Password, host, o.config.Database))
		}
		o.mutex.Lock()
		defer o.mutex.Unlock()
		var obj *xorm.EngineGroup
		obj, err := xorm.NewEngineGroup("mysql", dataSource, xorm.LeastConnPolicy())
		if err != nil {
			return nil, err
		}
		err = obj.DB().Ping()
		if err != nil {
			return nil, err
		}

		// 设置空闲连接池中的最大连接数
		obj.DB().SetMaxIdleConns(o.config.MaxIdleConns)
		// 设置数据库连接最大打开数
		obj.DB().SetMaxOpenConns(o.config.MaxOpenConns)
		// 设置可重用连接的最长时间，一定要小于mysql服务端的保持超时时间，否则可能会被服务端关闭
		obj.DB().SetConnMaxLifetime(o.config.MaxLifeTime)
		obj.SetMapper(names.GonicMapper{})
		// 非生产环境开启sql日志
		if strings.ToUpper(o.config.Env) == "DEV" || strings.ToUpper(o.config.Env) == "TEST" {
			obj.ShowSQL(true)
			obj.DB().SetConnMaxLifetime(1 * time.Minute)
		}
		dbMap[o.config.Database] = obj
		return &DB{obj}, nil
	}
}
