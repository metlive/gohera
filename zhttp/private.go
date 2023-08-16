package zhttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

const (
	TRACE_SPAN_ID        = "x-span-id"
	TRACE_RPC_ID_DEF     = "0"
	TRACE_CURRENT_RPC_ID = "trace_current_rpc_id"
)

// 设置链路追踪,trace_id相关
func (h *HTTPRequest) setTrace(ctx *gin.Context) *HTTPRequest {
	if ctx == nil {
		return h
	}
	prefixHeaders := ctx.Value("header")
	if prefixHeaders == nil {
		return h
	}
	headers, ok := prefixHeaders.(map[string]interface{})
	if !ok {
		return h
	}
	for k, v := range headers {
		if v1, ok := v.(string); ok {
			h.Request.Header.Set(k, v1)
			if strings.ToUpper(TRACE_SPAN_ID) == strings.ToUpper(k) {
				ctx.Set(TRACE_SPAN_ID, v1)
			}
		}
	}

	return h
}

// 设置链路追踪,设置rpc_id,需要先执行setTrace，设置TRACE_RPC_ID
func (h *HTTPRequest) setNextRpcId(ctx *gin.Context) *HTTPRequest {
	if ctx == nil {
		return h
	}
	rpcId := ctx.Value(TRACE_SPAN_ID)
	if rpcId == nil {
		h.Request.Header.Set(TRACE_SPAN_ID, TRACE_RPC_ID_DEF)
		return h
	}
	rpcIds, ok := rpcId.(string)
	if !ok {
		h.Request.Header.Set(TRACE_SPAN_ID, TRACE_RPC_ID_DEF)
		return h
	}
	cid := ctx.Value(TRACE_CURRENT_RPC_ID)
	cidInt, ok := cid.(int64)
	if !ok {
		cidInt = 1
	}
	rpcIds = rpcIds + "." + strconv.FormatInt(cidInt, 10)
	atomic.AddInt64(&cidInt, 1)
	ctx.Set(TRACE_CURRENT_RPC_ID, cidInt)

	h.Request.Header.Set(TRACE_SPAN_ID, rpcIds)

	return h
}

// 自定义Transport,后续可扩展
func (h *HTTPRequest) getTransport() http.RoundTripper {
	if h.transport == nil {
		return http.DefaultTransport
	}

	return h.transport
}

// 组装http Client
func (h *HTTPRequest) packClient() *http.Client {
	h.Client.Transport = h.getTransport()
	h.Client.Timeout = h.Timeout
	return h.Client
}

// 发起http请求,获取响应并设置对应的值
func (h *HTTPRequest) doRequest(method, reqUrl string) *HTTPRespone {
	h.Request.Method = method
	u, err := url.Parse(reqUrl)
	response := &HTTPRespone{}
	if err != nil {
		response.error = err
		return response
	}
	h.Request.URL = u
	client := h.packClient()

	resp, err := client.Do(h.Request)
	if err != nil {
		response.error = err
		return response
	}
	defer resp.Body.Close()
	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			response.error = err
			return response
		}
	} else {
		reader = resp.Body
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		response.error = err
		return response
	}
	response.responseCode = resp.StatusCode
	response.responseCookie = resp.Cookies()
	response.responseHeader = resp.Header
	response.bytes = body
	h.response = response

	return response
}

func (h *HTTPRequest) setBody(body []byte) *HTTPRequest {
	bf := bytes.NewBuffer(body)
	h.Request.Body = io.NopCloser(bf)
	h.Request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bf), nil
	}
	h.Request.ContentLength = int64(len(body))
	return h
}

// 自动添加referers
func (h *HTTPRequest) addReferer(ctx *gin.Context) *HTTPRequest {
	appName := os.Getenv("OCEAN_APP")
	value := ""
	if ctx != nil && ctx.Request != nil {
		value = "http://" + ctx.Request.Host + ctx.Request.RequestURI
	} else if appName != "" {
		value = "http://" + appName + "/cmd"
	}
	if value != "" {
		h.Request.Header.Add("referer", value)
	}

	return h
}
