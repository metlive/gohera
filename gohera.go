package gohera

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
)

var (
	httpHost string
	httpPort int
)

func StartupService(engine *gin.Engine) {

	httpHost = GetString("zhttp.host")
	httpPort = GetInt("zhttp.port")
	if httpPort == 0 {
		panic(errors.New("zhttp host or port is not valid"))
	}
	addr := httpHost + ":" + strconv.Itoa(httpPort)
	ac := make(chan error)
	go func() {
		fmt.Printf("服务启动，运行模式：%v，版本号：%s，进程号：%d , ip：%s", GetEnv(), "1.0.0", os.Getpid(), addr)
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
