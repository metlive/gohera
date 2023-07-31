/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package gohera

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 路由注册
func registerRouter() {
	Engine.GET("/healthz", healthCheck)
	//找不路由报错
	Engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "找不到你要的内容,URL:" + c.Request.Host + c.Request.RequestURI,
		})
		return
	})

	//找不到方法报错
	Engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    405,
			"message": "找不到该方法",
		})
		return
	})
}
