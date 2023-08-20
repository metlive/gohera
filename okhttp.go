package gohera

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	METHODGET       = "GET"
	METHODPOST      = "POST"
	METHODPUT       = "PUT"
	FORMCONTENTTYPE = "application/x-www-form-urlencoded"
	JSONCONTENTTYPE = "application/json"
)

type HTTPRequest struct {
	Client    *http.Client
	Request   *http.Request
	transport http.RoundTripper
	Timeout   time.Duration
	response  *HTTPRespone
	ctx       context.Context
}

type HTTPRespone struct {
	responseCode   int
	responseHeader http.Header
	responseCookie []*http.Cookie
	bytes          []byte
	error          error
}

func NewRequest() *HTTPRequest {
	return &HTTPRequest{
		Client: &http.Client{},
		Request: &http.Request{
			Header: make(http.Header),
		},
		Timeout: 3 * time.Second,
	}
}

// 获取响应的header
func (h *HTTPRequest) GetRespHeader() http.Header {
	if h != nil && h.response != nil {
		return h.response.responseHeader
	}
	return nil
}

// 获取响应的cookie
func (h *HTTPRequest) GetRespCookie() []*http.Cookie {
	if h != nil && h.response != nil {
		return h.response.responseCookie
	}
	return nil
}

// 获取响应的状态码
func (h *HTTPRequest) GetRespStatus() int {
	if h != nil && h.response != nil {
		return h.response.responseCode
	}
	return 0
}

func (h *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	h.transport = transport
	return h
}

// SetTimeOut 主动设置超时时间,默认3秒超时
func (h *HTTPRequest) SetTimeOut(Timeout int) *HTTPRequest {
	h.Timeout = time.Duration(Timeout) * time.Second
	return h
}

// Header 主动设置header头,可以覆盖之前的配置,批量添加
func (h *HTTPRequest) SetHeaders(header map[string]string) *HTTPRequest {
	if len(header) > 0 {
		for k, v := range header {
			h.Request.Header.Set(k, v)
		}
	}

	return h
}

// Header 主动设置header头,可以覆盖之前的配置，单个添加
func (h *HTTPRequest) SetHeader(k, v string) *HTTPRequest {
	if k != "" {
		h.Request.Header.Set(k, v)
	}

	return h
}

// cookie批量添加
func (h *HTTPRequest) SetCookies(cookies map[string]string) *HTTPRequest {
	if len(cookies) > 0 {
		for k, v := range cookies {
			h.Request.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}
	return h
}

// cookie 单个添加
func (h *HTTPRequest) SetCookie(k, v string) *HTTPRequest {
	if k != "" {
		h.Request.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	return h
}

// 添加referer
func (h *HTTPRequest) SetReferer(referer string) *HTTPRequest {
	h.Request.Header.Add("referer", referer)
	return h
}

func (h *HTTPRequest) GetCtx(ctx *gin.Context, reqUrl string) *HTTPRespone {
	h.Request.Header.Add("Content-Type", FORMCONTENTTYPE)
	resp := h.setTrace(ctx).setReferer(ctx).doRequest(ctx, METHODGET, reqUrl)
	return resp
}

func (h *HTTPRequest) PostCtx(ctx *gin.Context, reqUrl string, params map[string]string) *HTTPRespone {
	args := &url.Values{}
	for key, value := range params {
		args.Add(key, value)
	}
	h.Request.Header.Set("Content-Type", FORMCONTENTTYPE)
	resp := h.setTrace(ctx).setReferer(ctx).setBody([]byte(args.Encode())).doRequest(ctx, METHODPOST, reqUrl)
	return resp
}

func (h *HTTPRequest) JsonPostCtx(ctx *gin.Context, reqUrl string, params interface{}) *HTTPRespone {
	h.Request.Header.Set("Content-Type", JSONCONTENTTYPE)

	requestBody, err := json.Marshal(params)
	resp := &HTTPRespone{}
	if err != nil {
		resp.error = errors.New("json marshal fail")
		return resp
	}
	resp = h.setTrace(ctx).setReferer(ctx).setBody(requestBody).doRequest(ctx, METHODPOST, reqUrl)
	return resp
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
func (h *HTTPRequest) setReferer(ctx *gin.Context) *HTTPRequest {
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

// 设置链路追踪,trace_id相关
func (h *HTTPRequest) setTrace(ctx context.Context) *HTTPRequest {
	if ctx == nil {
		return h
	}
	var traceInfo *Trace
	var spanId = SpanIdDefault
	if ctxValue, ok := ctx.(*gin.Context); ok {
		traceInfo = ctxValue.MustGet(TraceCtx).(*Trace)
		indexArr := strings.Split(traceInfo.SpanId, ".")
		index, _ := strconv.Atoi(indexArr[len(indexArr)-1])
		spanId = traceInfo.SpanId + "." + strconv.FormatInt(int64(index)+1, 10)
		ctxValue.Set(TraceCtx, &Trace{
			TraceId: traceInfo.TraceId,
			SpanId:  spanId,
			UserId:  traceInfo.UserId,
			Method:  traceInfo.Method,
			Path:    ctxValue.Request.URL.Host + ctxValue.Request.URL.Path,
			Status:  ctxValue.Writer.Status(),
		})
	} else {
		traceInfo = ctx.Value(TraceCtx).(*Trace)
		traceInfo.Headers = make(map[string]any)
	}
	for k, v := range traceInfo.Headers {
		if v1, ok := v.(string); ok {
			h.Request.Header.Set(k, v1)
		}
	}
	h.Request.Header.Set(SpanId, spanId)
	return h
}

// 发起http请求,获取响应并设置对应的值
func (h *HTTPRequest) doRequest(ctx context.Context, method, reqUrl string) *HTTPRespone {
	h.Request.Method = method
	u, err := url.Parse(reqUrl)
	Infotf(ctx, "request %v %v", method, u)
	response := &HTTPRespone{}
	if err != nil {
		response.error = err
		return response
	}
	h.Request.URL = u
	if h.transport == nil {
		h.Client.Transport = http.DefaultTransport
	} else {
		h.Client.Transport = h.transport
	}
	h.Client.Timeout = h.Timeout

	resp, err := h.Client.Do(h.Request)
	if err != nil {
		response.error = err
		return response
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			response.error = err
			return
		}
	}(resp.Body)
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

// 信息输出
func (zr *HTTPRespone) Byte() ([]byte, error) {
	if zr.error != nil {
		return nil, zr.error
	}
	return zr.bytes, nil
}

func (zr *HTTPRespone) ToJSON(ret any) error {
	if zr.error != nil {
		return zr.error
	}
	if zr.bytes == nil {
		return errors.New("body empty")
	}
	err := json.Unmarshal(zr.bytes, ret)
	return err
}
