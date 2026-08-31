package gohera

import (
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metlive/gohera/validator"
)

// EngineOption 引擎初始化选项函数
type EngineOption func(*engineConfig)

type engineConfig struct {
	env string
}

// WithEnv 设置运行环境（dev / test / pre / prod），引擎据此决定 Gin 运行模式：
//
//	dev  → gin.DebugMode
//	test → gin.TestMode
//	其他 → gin.ReleaseMode
//
// 未设置时由 parseEnv() 自动决定。
func WithEnv(env string) EngineOption {
	return func(c *engineConfig) { c.env = env }
}

// resolveGinMode 根据环境标识解析 Gin 运行模式
func resolveGinMode(env string) string {
	switch env {
	case DeployEnvDev:
		return gin.DebugMode
	case DeployEnvTest:
		return gin.TestMode
	default:
		return gin.ReleaseMode
	}
}

// InitEngine 创建并配置 Gin 引擎（不依赖数据库/Redis/日志等完整初始化）。
//
// 可独立于 InitApp() 使用，适用于只需要 Gin 引擎能力（中间件、路由、参数校验）
// 而无需 MySQL/Redis/配置等全套基础设施的场景。
//
//	InitEngine()                          // 默认（由环境变量决定 mode）
//	InitEngine(WithEnv(gohera.DeployEnvDev))  // 手动指定 dev 环境 → DebugMode
//	InitEngine(WithEnv("prod"))              // 手动指定 prod → ReleaseMode
//
// 引擎配置内容：
//   - Gin mode 设置（option 优先，否则由环境决定）
//   - 链路追踪中间件 TraceContext
//   - 异常恢复中间件 HandlerRecovery（非 dev 环境自动启用）
//   - /healthz 健康检查、404/405 路由
//   - pprof 性能分析（由 config key "zhttp.pprof" 控制，需先加载配置）
//   - 数字不解析为 float64
//   - 注册自定义参数校验器（中文翻译、IPv4、日期格式等）
func InitEngine(opts ...EngineOption) *gin.Engine {
	cfg := &engineConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.env != "" {
		gin.SetMode(resolveGinMode(cfg.env))
	} else if appEnv != "" {
		// appEnv 已由 parseEnv() 设置，沿用环境决定的 mode
		updateGinMode()
	}
	// 否则沿用 Gin 默认值（DebugMode），不干预
	engine := gin.New()

	// 链路追踪上下文
	engine.Use(TraceContext())

	// 异常捕获（非 dev 环境启用，带完整堆栈）
	if !IsDev() {
		engine.Use(HandlerRecovery(true))
	}

	// 基础路由组（配置 http.context_path 时挂前缀）：框架路由（healthz/pprof）
	// 与业务路由（gohera.Router()）共用，业务侧无需感知前缀
	baseRouterGroup = ContextPathGroup(engine)

	// 注册默认路由 (healthz, 404, 405)
	registerRouter(engine)

	// pprof 性能分析（通过配置开关控制，与业务路由一致挂在 http.context_path 前缀下）
	if GetInt("zhttp.pprof") == 1 {
		if prefix := normalizeContextPath(GetString("http.context_path")); prefix != "" {
			pprof.Register(engine, strings.TrimPrefix(prefix, "/")+"/debug/pprof")
		} else {
			pprof.Register(engine)
		}
	}

	// 数字不要解析成 float64
	binding.EnableDecoderUseNumber = true
	// 注册自定义参数验证器
	binding.Validator = new(validator.DefaultValidator)

	return engine
}
