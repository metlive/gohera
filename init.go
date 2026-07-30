package gohera

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/metlive/gohera/logger"
	"github.com/metlive/gohera/mysql"
	"github.com/metlive/gohera/okhttp"
	"github.com/metlive/gohera/redis"
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

	// 初始化日志：从配置桥接到独立 logger 包（可覆盖此前仅控制台的 Init）
	appPath := GetDefaultString("log.path", DefaultLogPath)
	filePath := appPath + "/" + appName
	opts := logger.Options{
		FilePath: filePath,
		Project:  GetString("http.service"),
	}
	// 未配置 log.stdout 时显式关闭控制台，贴近生产默认只写文件
	if IsSet("log.stdout") {
		opts.EnableStdout = logger.Bool(GetBool("log.stdout"))
	} else {
		opts.EnableStdout = logger.Bool(false)
	}
	logger.Init(opts)

	// 桥接服务名到 okhttp（Referer 等）；三方未走 InitApp 时可自行 SetDefaultService
	okhttp.SetDefaultService(GetString("http.service"))

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
					imysql, err := mysql.New(conf)
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

	engine := InitEngine()
	return engine
}
