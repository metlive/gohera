/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package gohera

import (
	"context"
	"strings"

	"1v1.group/zlog"
	"github.com/gin-gonic/gin"
)

const (
	HeaderPrefix  = "x-trailer-"
	HeaderTraceId = "x-trace-id"
	HeaderRpcId   = "x-rpc-id"
)

func HandlerContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := &zlog.TraceEntry{
			TraceId: c.GetHeader(zlog.TRACE_HEADER_TRACE_ID),
			RpcId:   c.GetHeader(zlog.TRACE_HEADER_RPC_ID),
		}
		c.Set("trace", t)
		c.Set("header", getSpecialPrefixHeader(c))
		c.Next()
	}
}

func getSpecialPrefixHeader(ctx context.Context) map[string]interface{} {
	prefixHeaders := make(map[string]interface{})
	headers := ctx.(*gin.Context).Request.Header
	if headers == nil {
		return nil
	}
	for k, v := range headers {
		if strings.Contains(strings.ToLower(k), HeaderPrefix) || strings.ToLower(k) == HeaderTraceId || strings.ToLower(k) == HeaderRpcId {
			if len(v) == 0 {
				prefixHeaders[k] = ""
			} else {
				prefixHeaders[k] = v[0]
			}
		}
	}
	return prefixHeaders
}
