package gohera

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  any    `json:"result"`
}

func JsonError(c *gin.Context, code int, message ...string) {
	msg := GetMessage(DefaultErrorMsg)
	if len(message) <= 0 {
		msg = message[0]
	}
	c.JSON(http.StatusOK, newHttpResponse(code, msg, ""))
}

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

func (r *httpResponse) GetCode() int {
	return r.Code
}

func (r *httpResponse) GetMessage() string {
	return r.Message
}

func (r *httpResponse) GetResult() any {
	return r.Result
}

type responseBodyWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}
