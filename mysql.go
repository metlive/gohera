package gohera

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
	Trace        bool          // 追踪调试
	MaxLifeTime  time.Duration // 设置连接可以被重新使用的最大时间量
	MaxOpenConns int           // 设置打开连接到数据库的最大数量
	MaxIdleConns int           // 设置空闲连接池中的最大连接数
	User         string        //用户名
	Password     string        //密码
	Host         string        //数据库地址
	Port         int           //端口
	Database     string        //连接那个数据库
}

type DB struct {
	*xorm.EngineGroup
}

var instance *ConnectPool
var once sync.Once

func NewMysql() *ConnectPool {
	once.Do(func() {
		instance = &ConnectPool{}
	})
	return instance
}

// 变量初始化
type ConnectPool struct {
	Mutex sync.Mutex
}

func (o *ConnectPool) initPool(config *Config) (*DB, error) {
	var dataSource []string
	hosts := strings.Split(config.Host, ",")
	for _, host := range hosts {
		dataSource = append(dataSource, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", config.User, config.Password, host, config.Database))
	}
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	obj, err := xorm.NewEngineGroup("mysql", dataSource, xorm.LeastConnPolicy())
	if err != nil {
		return nil, err
	}
	err = obj.Ping()
	if err != nil {
		return nil, err
	}

	// 设置空闲连接池中的最大连接数
	obj.SetMaxIdleConns(config.MaxIdleConns)
	// 设置数据库连接最大打开数
	obj.SetMaxOpenConns(config.MaxOpenConns)
	// 设置可重用连接的最长时间，一定要小于mysql服务端的保持超时时间，否则可能会被服务端关闭
	obj.SetConnMaxLifetime(config.MaxLifeTime)
	obj.SetMapper(names.GonicMapper{})

	return &DB{obj}, nil
}
