package gohera

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
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
			spanID = "1"
		} else {
			indexArr := strings.Split(spanID, ".")
			index, _ := strconv.Atoi(indexArr[len(indexArr)-1])
			spanID = spanID + "." + strconv.Itoa(index+1)
		}
		userId := c.GetInt(UserId)
		t := &Trace{
			TraceId: traceID,
			SpanId:  spanID,
			UserId:  userId,
			Method:  c.Request.Method,
			Path:    c.Request.URL.Host + c.Request.URL.Path,
			Status:  c.Writer.Status(),
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
