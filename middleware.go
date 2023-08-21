package gohera

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
)

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
			Headers: func(headers map[string][]string) map[string]any {
				prefixHeaders := make(map[string]any)
				for k, v := range headers {
					if len(v) == 0 {
						prefixHeaders[k] = ""
					} else {
						prefixHeaders[k] = v[0]
					}
				}
				return prefixHeaders
			}(c.Request.Header),
		}
		c.Set(TraceCtx, t)
		c.Next()
	}
}

func RequestContext() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func RecoveryContext() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
