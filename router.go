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
