package gohera

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"os"
)

var env = flag.String("env", DeployEnvDev, "The environment for app run")

const DefaultLogPath = "/var/log/trace"

func InitApp() error {
	flag.Parse()

	// 解析环境变量
	err := parseEnv(*env)
	if err != nil {
		return err
	}

	// 初始化应用配置
	err = initAppConfig()
	if err != nil {
		return err
	}

	// 初始化日志处理器
	appPath := GetString("log.path")
	if appPath == "" {
		appPath = DefaultLogPath
	}
	initLoggerPool(loggerConfig{
		FilePath:   appPath + "/" + appName,
		MaxSize:    0,
		MaxBackups: 0,
		Compress:   false,
		Mode:       "",
	})

	// mysql初始化
	if IsSet("mysql") {
		dbList := GetStringMap("mysql")
		for key := range dbList {
			if IsSet("mysql." + key) {
				var conf = new(dbConfig)
				if err = config.UnmarshalKey("mysql."+key, &conf); err != nil {
					panic(fmt.Errorf("unable to decode dbConfig struct：  %s \n pid:%d", err, os.Getpid()))
				}
				Mysql[key] = func(conf *dbConfig) *DB {
					mysql, err := NewMysql().initPool(conf)
					if err != nil {
						return nil
					}
					return mysql
				}(conf)
				fmt.Println("==="+key+"====", *Mysql[key], Mysql[key].DataSourceName())
			}
		}
		//mysqlParams := new(dbConfig)
		//if err = config.UnmarshalKey("mysql", &mysqlParams); err != nil {
		//	panic(fmt.Errorf("unable to decode dbConfig struct：  %s \n pid:%d", err, os.Getpid()))
		//}

	}
	//
	//// redis初始化
	//if IsSet("redis") {
	//	redisConfig := GetStringMap("redis")
	//	Redis, err = redis.New(redisConfig)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func geMySql(conf *dbConfig) *DB {
	mysql, err := NewMysql().initPool(conf)
	if err != nil {
		return nil
	}
	return mysql
}

func InitHttpServer() {
	// init engine
	Engine = gin.New()
	registerMiddleware()
	registerRouter()

	// 是否需要开启pprof
	prof := GetInt("http.pprof")
	if prof == 1 {
		pprof.Register(Engine)
	}

	// 数字不要解析成float64
	binding.EnableDecoderUseNumber = true
}
