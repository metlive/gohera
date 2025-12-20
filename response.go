package gohera

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  any    `json:"data"`
}

var contexts *gin.Context

// JsonError 返回 JSON 格式的错误响应
func JsonError(c *gin.Context, code int, message ...string) {
	msg := GetMessage(DefaultErrorMsg)
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusOK, newHttpResponse(code, msg, ""))
}

// JsonSuccess 返回 JSON 格式的成功响应
func JsonSuccess(c *gin.Context, result any) {
	c.JSON(http.StatusOK, newHttpResponse(Success, "", result))
}

// 终止请求
func JsonAbort(c *gin.Context, errCode int, message string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, newHttpResponse(errCode, message, ""))
}

func newHttpResponse(errCode int, message string, result any) *httpResponse {
	rsp := &httpResponse{}

	rsp.Code = errCode
	if message != "" {
		rsp.Message = message
	} else {
		rsp.Message = GetMessage(errCode)
	}
	rsp.Result = result

	return rsp
}

// GetCode 获取响应码
func (r *httpResponse) GetCode() int {
	return r.Code
}

// GetMessage 获取响应消息
func (r *httpResponse) GetMessage() string {
	return r.Message
}

// GetResult 获取响应结果
func (r *httpResponse) GetResult() any {
	return r.Result
}

type responseBodyWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

// Write 实现 gin.ResponseWriter 接口，拦截写入内容
func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}
