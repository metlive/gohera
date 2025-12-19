package gohera

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metlive/gohera/mysql"
	"github.com/metlive/gohera/redis"
	"github.com/metlive/gohera/validator"
)

var env = flag.String("env", DeployEnvDev, "The environment for app run")

const DefaultLogPath = "/var/log/trace"

// InitApp 初始化应用
// 解析环境变量、配置文件、日志、数据库（MySQL/Redis）、PProf 和验证器，并返回 Gin 引擎
func InitApp() (router *gin.Engine) {
	flag.Parse()

	// 解析环境变量
	err := parseEnv(*env)
	if err != nil {
		panic(fmt.Errorf("env parse fail ：  %s \n", err))
	}

	// 初始化应用配置
	err = initAppConfig()
	if err != nil {
		panic(fmt.Errorf("init config fail ：  %s \n", err))
	}

	// 初始化日志处理器
	appPath := GetDefaultString("log.path", DefaultLogPath)
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
				conf := new(mysql.Config)
				if err = UnmarshalKey("mysql."+key, conf); err != nil {
					panic(fmt.Errorf("unable to decode dbConfig struct：  %s \n pid:%d", err, os.Getpid()))
				}
				Mysql[key] = func(conf *mysql.Config) *mysql.DB {
					conf.Env = GetEnv()
					imysql, err := mysql.InitOnce(conf).Connect()
					if err != nil {
						panic(fmt.Errorf("unable to connect fail ：  %s \n", err))
					}
					return imysql
				}(conf)
			}
		}
	}

	// redis初始化
	if IsSet("redis") {
		conf := new(redis.Config)
		if err = UnmarshalKey("redis", conf); err != nil {
			panic(fmt.Errorf("unable to decode dbConfig struct：  %s \n pid:%d", err, os.Getpid()))
		}
		Redis, err = redis.New(conf)
		if err != nil {
			panic(fmt.Errorf("unable to connect fail ：  %s \n", err))
		}
	}

	engine := gin.New()
	// 初始化上下文
	engine.Use(TraceContext())
	// 异常捕获
	if GetEnv() != DeployEnvDev {
		engine.Use(HandlerRecovery(true))
	}
	// 记录请求日志
	registerRouter(engine)

	// 是否需要开启pprof
	prof := GetInt("zhttp.pprof")
	if prof == 1 {
		pprof.Register(engine)
	}

	// 数字不要解析成float64
	binding.EnableDecoderUseNumber = true
	// 注册自定义参数验证
	binding.Validator = new(validator.DefaultValidator)
	return engine
}
