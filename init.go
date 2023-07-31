package gohera

import (
	"flag"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		FilePath:   appPath + "/" + appName + "_%Y%m%d.log",
		MaxSize:    0,
		MaxBackups: 0,
		Compress:   false,
		Mode:       "",
	})

	// mysql初始化
	//if IsSet("mysql") {
	//	mysqlConf := GetStringMap("mysql")
	//	Mysql, err = mysql.New(mysqlConf)
	//	if err != nil {
	//		return err
	//	}
	//	// 非生产环境开启sql日志
	//	if GetEnv() == DeployEnvDev || GetEnv() == DeployEnvTest {
	//		Mysql.ShowSQL(true)
	//		if GetEnv() == DeployEnvTest {
	//			Mysql.SetConnMaxLifetime(1 * time.Minute)
	//		}
	//	}
	//}
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
