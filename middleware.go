package gohera

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceContext 生成链路追踪 ID 的中间件
// 会检查 Request Header 中的 TraceId, SpanId，如果不存在则生成新的
func TraceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceId)
		spanID := c.GetHeader(SpanId)
		if traceID == "" {
			if c.GetHeader("HTTP_TRACE_ID") != "" {
				traceID = c.GetHeader("HTTP_TRACE_ID")
			} else if c.GetHeader("TRACE_ID") != "" {
				traceID = c.GetHeader("TRACE_ID")
			} else {
				traceID = strings.ReplaceAll(uuid.NewString(), "-", "")
			}
		}
		if spanID == "" {
			spanID = SpanIdDefault
		}
		t := &Trace{
			TraceId: traceID,
			SpanId:  spanID,
			UserId:  c.GetInt(UserId),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Host + c.Request.URL.Path,
			Status:  c.Writer.Status(),
			Headers: getHeader(c.Request.Header),
		}
		c.Set(TraceCtx, t)
		// 将 Trace 信息注入到 Request 的 Context 中，以便在非 Gin 环境下（如 Service 层）也能获取
		ctx := context.WithValue(c.Request.Context(), TraceCtx, t)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// CorsContext 处理跨域请求 (CORS) 的中间件
func CorsContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		// 跨域自定义header ，不设置自动加入，在下面手动设置
		cosHeader := c.GetHeader("Access-Control-Request-Headers")
		cosHeaders := strings.Split(cosHeader, ",")
		headerKeys = append(headerKeys, cosHeaders...)

		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Headers", headerStr)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// 放行所有OPTIONS方法
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatusJSON(http.StatusNoContent, "WeclassRoom Request Options")
		}
		c.Next()
	}
}

// RecoveryContext 空实现的 Recovery 中间件 (预留)
func RecoveryContext() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}
