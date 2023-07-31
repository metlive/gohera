/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

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
