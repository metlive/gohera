package gohera

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HTTPRequest struct {
	client    *http.Client
	request   *http.Request
	transport http.RoundTripper
	timeout   time.Duration
	response  *HTTPRespone
	ctx       context.Context
	retries   int
	params    map[string][]string
	url       string
}

type HTTPRespone struct {
	response       *http.Response
	responseCode   int
	responseHeader http.Header
	responseCookie []*http.Cookie
	bytes          []byte
	error          error
}

func NewRequest() *HTTPRequest {
	return &HTTPRequest{
		client: &http.Client{},
		request: &http.Request{
			Header: make(http.Header),
		},

		params:  make(map[string][]string),
		timeout: 3 * time.Second,
		retries: 1,
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

// sets the schema's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (h *HTTPRequest) SetBasicAuth(username, password string) *HTTPRequest {
	h.request.SetBasicAuth(username, password)
	return h
}

func (h *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	h.transport = transport
	return h
}

// 主动设置超时时间,默认3秒超时
func (h *HTTPRequest) SetTimeOut(timeout int) *HTTPRequest {
	h.timeout = time.Duration(timeout) * time.Second
	return h
}

// Header 主动设置header头,可以覆盖之前的配置,批量添加
func (h *HTTPRequest) SetHeaders(header map[string]string) *HTTPRequest {
	if len(header) > 0 {
		for k, v := range header {
			h.request.Header.Set(k, v)
		}
	}
	return h
}

// Header 主动设置header头,可以覆盖之前的配置，单个添加
func (h *HTTPRequest) SetHeader(k, v string) *HTTPRequest {
	if k != "" {
		h.request.Header.Set(k, v)
	}
	return h
}

// cookie批量添加
func (h *HTTPRequest) SetCookies(cookies map[string]string) *HTTPRequest {
	if len(cookies) > 0 {
		for k, v := range cookies {
			h.request.AddCookie(&http.Cookie{
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
		h.request.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	return h
}

// 添加referer
func (h *HTTPRequest) SetReferer(referer string) *HTTPRequest {
	h.request.Header.Add("referer", referer)
	return h
}

// default is 0 means no retried.
// -1 means retried forever.
// others means retried times.
func (h *HTTPRequest) SetRetries(times int) *HTTPRequest {
	h.retries = times
	return h
}

// Param adds query param in to schema.
// params build query string as ?key1=value1&key2=value2...
func (h *HTTPRequest) SetParam(key string, value any) *HTTPRequest {
	if param, ok := h.params[key]; ok {
		h.params[key] = append(param, fmt.Sprintf("%v", value))
	} else {
		h.params[key] = []string{fmt.Sprintf("%v", value)}
	}
	return h
}

func (h *HTTPRequest) GetCtx(ctx context.Context, reqUrl string) *HTTPRespone {
	h.request.Header.Add("Content-Type", FormContentType)
	if len(h.params) > 0 {
		var paramBody string
		var buf bytes.Buffer
		for k, v := range h.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
		if strings.Contains(reqUrl, "?") {
			reqUrl += "&" + paramBody
		} else {
			reqUrl = reqUrl + "?" + paramBody
		}
	}
	resp := h.setTrace(ctx).setReferer(ctx).doRequest(http.MethodGet, reqUrl)
	return resp
}

func (h *HTTPRequest) DeleteCtx(ctx *gin.Context, reqUrl string) *HTTPRespone {
	h.request.Header.Add("Content-Type", FormContentType)
	resp := h.setTrace(ctx).setReferer(ctx).doRequest(http.MethodDelete, reqUrl)
	return resp
}

func (h *HTTPRequest) PostFormCtx(ctx *gin.Context, reqUrl string, params map[string]any) *HTTPRespone {
	args := &url.Values{}
	for key, value := range params {
		args.Add(key, fmt.Sprintf("%v", value))
	}
	h.request.Header.Set("Content-Type", FormContentType)
	resp := h.setTrace(ctx).setReferer(ctx).setBody([]byte(args.Encode())).doRequest(http.MethodPost, reqUrl)
	return resp
}

func (h *HTTPRequest) PostJsonCtx(ctx *gin.Context, reqUrl string, params any) *HTTPRespone {
	h.request.Header.Set("Content-Type", JsonContentType)
	requestBody, err := json.Marshal(params)
	resp := &HTTPRespone{}
	if err != nil {
		resp.error = errors.New("json marshal fail")
		return resp
	}
	resp = h.setTrace(ctx).setReferer(ctx).setBody(requestBody).doRequest(http.MethodPost, reqUrl)
	return resp
}

func (h *HTTPRequest) setBody(body []byte) *HTTPRequest {
	bf := bytes.NewBuffer(body)
	h.request.Body = io.NopCloser(bf)
	h.request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bf), nil
	}
	h.request.ContentLength = int64(len(body))
	return h
}

// 自动添加referers
func (h *HTTPRequest) setReferer(ctx context.Context) *HTTPRequest {
	appName := os.Getenv("OCEAN_APP")
	value := ""
	if ctx != nil && h.request != nil {
		value = "http://" + h.request.Host + h.request.RequestURI
	} else if appName != "" {
		value = "http://" + appName + "/cmd"
	}
	if value != "" {
		h.request.Header.Add("referer", value)
	}
	return h
}

// 设置链路追踪,trace_id相关
func (h *HTTPRequest) setTrace(cx context.Context) *HTTPRequest {
	if cx == nil {
		cx = context.Background()
	}
	traceInfo := new(Trace)
	spanId := SpanIdDefault
	if ctx, ok := cx.(*gin.Context); ok {
		traceInfo = ctx.MustGet(TraceCtx).(*Trace)
		if traceInfo.SpanId == "" {
			traceInfo.SpanId = SpanIdDefault
		}
		indexArr := strings.Split(traceInfo.SpanId, ".")
		index, _ := strconv.Atoi(indexArr[len(indexArr)-1])
		spanId = traceInfo.SpanId + "." + strconv.FormatInt(int64(index)+1, 10)
		ctx.Set(TraceCtx, &Trace{
			TraceId: traceInfo.TraceId,
			SpanId:  spanId,
			UserId:  traceInfo.UserId,
			Method:  traceInfo.Method,
			Path:    ctx.Request.URL.Host + ctx.Request.URL.Path,
			Status:  ctx.Writer.Status(),
			Headers: getHeader(h.request.Header),
		})
	} else {
		if h.request.Header.Get(SpanId) != "" {
			traceInfo.SpanId = h.request.Header.Get(SpanId)
		}
		if traceInfo.SpanId == "" {
			traceInfo.SpanId = SpanIdDefault
		}
		indexArr := strings.Split(traceInfo.SpanId, ".")
		index, _ := strconv.Atoi(indexArr[len(indexArr)-1])
		spanId = traceInfo.SpanId + "." + strconv.FormatInt(int64(index)+1, 10)
		traceInfo = &Trace{
			TraceId: strings.ReplaceAll(uuid.NewString(), "-", ""),
			SpanId:  spanId,
			UserId:  0,
			Method:  h.request.Method,
			Path:    h.url,
			Status:  200,
			Headers: getHeader(h.request.Header),
		}
		context.WithValue(cx, TraceCtx, traceInfo)
	}
	for k, v := range traceInfo.Headers {
		if v1, ok := v.(string); ok {
			h.request.Header.Set(k, v1)
		}
	}
	h.request.Header.Set(SpanId, spanId)
	return h
}

// 发起http请求,获取响应并设置对应的值
func (h *HTTPRequest) doRequest(method, reqUrl string) *HTTPRespone {
	h.request.Method = method
	u, err := url.Parse(reqUrl)
	Infotf(h.ctx, "request %v: %v", method, u)
	response := &HTTPRespone{}
	if err != nil {
		response.error = err
		return response
	}
	h.request.URL = u
	if h.transport == nil {
		h.client.Transport = http.DefaultTransport
	} else {
		h.client.Transport = h.transport
	}
	h.client.Timeout = h.timeout
	// 请求测试次数
	var resp *http.Response
	for i := 0; h.retries == -1 || i <= h.retries; i++ {
		resp, err = h.client.Do(h.request)
		if err == nil {
			break
		}
	}

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
	response.response = resp
	h.response = response

	return response
}

// 信息输出
func (zr *HTTPRespone) Bytes() ([]byte, error) {
	if zr.error != nil {
		return nil, zr.error
	}
	return zr.bytes, nil
}

func (zr *HTTPRespone) String() (string, error) {
	if zr.error != nil {
		return "", zr.error
	}
	return string(zr.bytes), nil
}

func (zr *HTTPRespone) Response() (*http.Response, error) {
	return zr.Response()
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
