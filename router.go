package gohera

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerRouter 注册默认路由 (Healthz, 404, 405)
func registerRouter(engine *gin.Engine) {
	engine.GET("/healthz", healthCheck)
	//找不路由报错
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, newHttpResponse(http.StatusNotFound, "找不到你要的内容,URL:"+c.Request.Host+c.Request.RequestURI, ""))
		return
	})

	//找不到方法报错
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, newHttpResponse(http.StatusMethodNotAllowed, "找不到该方法", ""))
		return
	})
}
