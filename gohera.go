package gohera

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/gin-gonic/gin"
)

// StartupOption 启动选项函数
type StartupOption func(*startupConfig)

type startupConfig struct {
	host string
	port int
}

// WithHost 指定监听地址（host），传空字符串则回退到配置文件
func WithHost(host string) StartupOption {
	return func(c *startupConfig) { c.host = host }
}

// WithPort 指定监听端口，传 0 则回退到配置文件
func WithPort(port int) StartupOption {
	return func(c *startupConfig) { c.port = port }
}

// StartupService 启动 HTTP 服务
// 根据配置启动 Gin 引擎，并处理平滑退出信号。
//
// 默认从配置文件读取 http.host / http.port，也可以通过 WithHost / WithPort 覆盖：
//
//	StartupService(engine)                                    // 全部从配置文件读取
//	StartupService(engine, WithHost("0.0.0.0"))               // host 手动指定，port 从配置读取
//	StartupService(engine, WithPort(8080))                    // port 手动指定，host 从配置读取
//	StartupService(engine, WithHost("127.0.0.1"), WithPort(9090)) // 全部手动指定
func StartupService(engine *gin.Engine, opts ...StartupOption) {
	cfg := &startupConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	host := cfg.host
	if host == "" {
		host = GetString("http.host")
	}

	port := cfg.port
	if port == 0 {
		port = GetInt("http.port")
	}
	if port == 0 {
		panic(errors.New("http port is not valid"))
	}

	addr := host + ":" + strconv.Itoa(port)
	ac := make(chan error)
	go func() {
		fmt.Printf("服务启动，运行模式：%v，版本号：%s，进程号：%d , ip：%s", GetEnv(), GetAppVersion(), os.Getpid(), addr)
		fmt.Println("")
		err := engine.Run(addr)
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("监听HTTP服务: %v", err.Error())
			ac <- err
		}
	}()
	var state int32 = 1
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case err := <-ac:
		if err != nil && atomic.LoadInt32(&state) == 1 {
			Error(context.Background(), "监听HTTP服务发生错误: %v", err.Error())
			panic(fmt.Sprintf("监听HTTP服务发生错误: %v", err.Error()))
		}
	case sig := <-quit:
		atomic.StoreInt32(&state, 0)
		fmt.Printf("获取到退出信号: %v  pid %d", sig.String(), os.Getpid())
	}
}
