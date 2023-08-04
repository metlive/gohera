package gohera

import (
	"strings"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func HandlerContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		//ctx := context.Background()
		traceID := c.GetHeader(TraceHeaderTraceId)
		if traceID == "" {
			if c.GetHeader("HTTP_TRACE_ID") != "" {
				traceID = c.GetHeader("HTTP_TRACE_ID")
			} else if c.GetHeader("TRACE_ID") != "" {
				traceID = c.GetHeader("TRACE_ID")
			} else {
				traceID = strings.ReplaceAll(uuid.NewString(), "-", "")
			}
		}
		userId := c.GetInt(TraceHeaderUserId)
		t := &Trace{
			TraceId: traceID,
			SpanId:  "",
			UserId:  userId,
			Method:  c.Request.Method,
			Path:    c.Request.URL.Host + c.Request.URL.Path,
			Status:  c.Writer.Status(),
		}
		c.Set(TraceCtx, t)
		c.Next()
	}
}
