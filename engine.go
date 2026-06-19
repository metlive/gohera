package gohera

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metlive/gohera/validator"
)

// InitEngine 创建并配置 Gin 引擎（不依赖数据库/Redis/日志等完整初始化）。
//
// 可独立于 InitApp() 使用，适用于只需要 Gin 引擎能力（中间件、路由、参数校验）
// 而无需 MySQL/Redis/配置等全套基础设施的场景。
//
// 引擎配置内容：
//   - 链路追踪中间件 TraceContext
//   - 异常恢复中间件 HandlerRecovery（非 dev 环境自动启用）
//   - /healthz 健康检查、404/405 路由
//   - pprof 性能分析（由 config key "zhttp.pprof" 控制，需先加载配置）
//   - 数字不解析为 float64
//   - 注册自定义参数校验器（中文翻译、IPv4、日期格式等）
func InitEngine() *gin.Engine {
	engine := gin.New()

	// 链路追踪上下文
	engine.Use(TraceContext())

	// 异常捕获（非 dev 环境启用，带完整堆栈）
	if !IsDev() {
		engine.Use(HandlerRecovery(true))
	}

	// 注册默认路由 (healthz, 404, 405)
	registerRouter(engine)

	// pprof 性能分析（通过配置开关控制）
	if GetInt("zhttp.pprof") == 1 {
		pprof.Register(engine)
	}

	// 数字不要解析成 float64
	binding.EnableDecoderUseNumber = true
	// 注册自定义参数验证器
	binding.Validator = new(validator.DefaultValidator)

	return engine
}
