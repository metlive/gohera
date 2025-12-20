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

var (
	httpHost string
	httpPort int
)

// StartupService 启动 HTTP 服务
// 根据配置启动 Gin 引擎，并处理平滑退出信号
func StartupService(engine *gin.Engine) {
	httpHost = GetString("http.host")
	httpPort = GetInt("http.port")
	if httpPort == 0 {
		panic(errors.New("http host or port is not valid"))
	}
	addr := httpHost + ":" + strconv.Itoa(httpPort)
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
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
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
