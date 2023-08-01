package gohera

func registerMiddleware() {
	// 初始化上下文
	Engine.Use(HandlerContext())

	// 异常捕获
	if GetEnv() != DeployEnvDev {
		Engine.Use(HandlerRecovery(true))
	}

	// 记录请求日志
	//Engine.Use(HandleAppAccessLog())
}
