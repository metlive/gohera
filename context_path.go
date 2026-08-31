package gohera

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextPathGroup 返回挂在 http.context_path 前缀下的基础路由组：
// 配置了前缀（如 message-dispatcher）时为 engine.Group("/message-dispatcher")，
// 未配置或留空（"/"）时为根组，行为与直接在 engine 上注册一致。
//
// 业务路由统一注册在该组下即自动获得接口前缀（如 /message-dispatcher/api/v1/*），
// 根路径不再注册对应路由。前缀在进程启动时的路由注册阶段决定，修改配置需重启。
//
// 常规用法无需自行调用：InitEngine 已设置框架基础路由组，业务路由经 gohera.Router() 注册。
func ContextPathGroup(engine *gin.Engine) *gin.RouterGroup {
	prefix := normalizeContextPath(GetString("http.context_path"))
	if prefix == "" {
		return engine.Group("/")
	}
	return engine.Group(prefix)
}

// baseRouterGroup 由 InitEngine 设置的基础路由组，Router() 返回它。
var baseRouterGroup *gin.RouterGroup

// Router 返回框架基础路由组（配置 http.context_path 时挂前缀，未配置时为根组）。
// 框架自身路由（healthz、pprof）与业务路由共用该组，业务侧在其下注册即自动
// 获得接口前缀，无需感知配置：
//
//	engine := gohera.InitApp()
//	gohera.Router().GET("/health", ...)
//	v1 := gohera.Router().Group("/api/v1")
//
// 须在 InitApp / InitEngine 之后调用，否则 panic。
func Router() *gin.RouterGroup {
	if baseRouterGroup == nil {
		panic("gohera.Router() called before InitApp()/InitEngine()")
	}
	return baseRouterGroup
}

// normalizeContextPath 归一化为以 / 开头、不以 / 结尾；空值与 "/" 视为无前缀。
func normalizeContextPath(raw string) string {
	p := strings.Trim(strings.TrimSpace(raw), "/")
	if p == "" {
		return ""
	}
	return "/" + p
}
