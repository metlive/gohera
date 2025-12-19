package gohera

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	defaultHTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   20,
		},
	}
)

type HTTPRequest struct {
	client    *http.Client
	transport http.RoundTripper
	header    http.Header
	timeout   time.Duration
	response  *HTTPRespone
	ctx       context.Context
	retries   int
	params    url.Values
	url       string
	body      []byte
	method    string
}

type HTTPRespone struct {
	response       *http.Response
	responseCode   int
	responseHeader http.Header
	responseCookie []*http.Cookie
	bytes          []byte
	Error          error // 改为导出字段或保持兼容性，但计划中提到导出或提供更好访问
}

// NewRequest 创建一个新的 HTTPRequest 实例
// 默认 3秒超时，重试 1 次
func NewRequest() *HTTPRequest {
	return &HTTPRequest{
		client:  defaultHTTPClient,
		header:  make(http.Header),
		params:  make(url.Values),
		timeout: 3 * time.Second,
		retries: 1,
	}
}

// GetRespHeader 获取响应的 Header
func (h *HTTPRequest) GetRespHeader() http.Header {
	if h != nil && h.response != nil {
		return h.response.responseHeader
	}
	return nil
}

// GetRespCookie 获取响应的 Cookie
func (h *HTTPRequest) GetRespCookie() []*http.Cookie {
	if h != nil && h.response != nil {
		return h.response.responseCookie
	}
	return nil
}

// GetRespStatus 获取响应的状态码
func (h *HTTPRequest) GetRespStatus() int {
	if h != nil && h.response != nil {
		return h.response.responseCode
	}
	return 0
}

// SetBasicAuth 设置 HTTP Basic Auth 认证头
func (h *HTTPRequest) SetBasicAuth(username, password string) *HTTPRequest {
	h.header.Set("Authorization", "Basic "+basicAuth(username, password))
	return h
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// SetTransport 设置自定义的 http.RoundTripper
func (h *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	h.transport = transport
	return h
}

// SetTimeOut 设置请求超时时间 (默认3秒)
func (h *HTTPRequest) SetTimeOut(timeout int) *HTTPRequest {
	h.timeout = time.Duration(timeout) * time.Second
	return h
}

// SetHeaders 批量设置请求头 (覆盖现有同名 Header)
func (h *HTTPRequest) SetHeaders(header map[string]string) *HTTPRequest {
	if len(header) > 0 {
		for k, v := range header {
			h.header.Set(k, v)
		}
	}
	return h
}

// SetHeader 设置单个请求头 (覆盖现有同名 Header)
func (h *HTTPRequest) SetHeader(k, v string) *HTTPRequest {
	if k != "" {
		h.header.Set(k, v)
	}
	return h
}

// SetCookies 批量添加 Cookie
func (h *HTTPRequest) SetCookies(cookies map[string]string) *HTTPRequest {
	if len(cookies) > 0 {
		for k, v := range cookies {
			h.SetCookie(k, v)
		}
	}
	return h
}

// SetCookie 添加单个 Cookie
func (h *HTTPRequest) SetCookie(k, v string) *HTTPRequest {
	if k != "" {
		h.header.Add("Cookie", (&http.Cookie{Name: k, Value: v}).String())
	}
	return h
}

// SetReferer 设置 Referer 头
func (h *HTTPRequest) SetReferer(referer string) *HTTPRequest {
	h.header.Add("Referer", referer)
	return h
}

// SetRetries 设置重试次数 (0: 不重试, -1: 无限重试, >0: 重试次数)
func (h *HTTPRequest) SetRetries(times int) *HTTPRequest {
	h.retries = times
	return h
}

// SetParam 添加查询参数
func (h *HTTPRequest) SetParam(key string, value any) *HTTPRequest {
	h.params.Add(key, fmt.Sprintf("%v", value))
	return h
}

// Get 发起 GET 请求
func (h *HTTPRequest) Get(reqUrl string) *HTTPRespone {
	ctx := context.Background()
	return h.GetCtx(ctx, reqUrl)
}

// GetCtx 发起带 Context 的 GET 请求
func (h *HTTPRequest) GetCtx(ctx context.Context, reqUrl string) *HTTPRespone {
	h.header.Set("Content-Type", FormContentType)
	h.url = reqUrl
	h.method = http.MethodGet
	return h.doRequest(ctx)
}

// DeleteCtx 发起带 Context 的 DELETE 请求
func (h *HTTPRequest) DeleteCtx(ctx context.Context, reqUrl string) *HTTPRespone {
	h.header.Set("Content-Type", FormContentType)
	h.url = reqUrl
	h.method = http.MethodDelete
	return h.doRequest(ctx)
}

// PostFormCtx 发起带 Context 的 POST Form 请求
func (h *HTTPRequest) PostFormCtx(ctx context.Context, reqUrl string, params map[string]any) *HTTPRespone {
	args := url.Values{}
	for key, value := range params {
		args.Add(key, fmt.Sprintf("%v", value))
	}
	h.header.Set("Content-Type", FormContentType)
	h.url = reqUrl
	h.method = http.MethodPost
	h.body = []byte(args.Encode())
	return h.doRequest(ctx)
}

// PostJsonCtx 发起带 Context 的 POST JSON 请求
func (h *HTTPRequest) PostJsonCtx(ctx context.Context, reqUrl string, params any) *HTTPRespone {
	h.header.Set("Content-Type", JsonContentType)
	requestBody, err := json.Marshal(params)
	if err != nil {
		return &HTTPRespone{Error: errors.New("json marshal fail: " + err.Error())}
	}
	h.url = reqUrl
	h.method = http.MethodPost
	h.body = requestBody
	return h.doRequest(ctx)
}

func (h *HTTPRequest) setBody(req *http.Request) {
	if len(h.body) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(h.body))
		req.ContentLength = int64(len(h.body))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(h.body)), nil
		}
	}
}

// 自动添加referers
func (h *HTTPRequest) setReferer(ctx context.Context, req *http.Request) {
	if h.header.Get("Referer") != "" {
		return
	}
	appName := GetString("http.service")
	value := ""
	if req.URL != nil {
		value = `http://` + req.Host + req.URL.RequestURI()
	} else if appName != "" {
		value = `http://` + appName + "/cmd"
	}
	if value != "" {
		req.Header.Set("Referer", value)
	}
}

// 设置链路追踪,trace_id相关
func (h *HTTPRequest) setTrace(cx context.Context, req *http.Request) context.Context {
	if cx == nil {
		cx = context.Background()
	}
	traceInfo := new(Trace)
	spanId := SpanIdDefault

	if gCtx, ok := cx.(*gin.Context); ok {
		if val, exists := gCtx.Get(TraceCtx); exists {
			if t, ok := val.(*Trace); ok {
				traceInfo = t
			}
		}
		if traceInfo.SpanId == "" {
			traceInfo.SpanId = SpanIdDefault
		}
		indexArr := strings.Split(traceInfo.SpanId, ".")
		index, _ := strconv.Atoi(indexArr[len(indexArr)-1])
		spanId = traceInfo.SpanId + "." + strconv.FormatInt(int64(index)+1, 10)

		newTrace := &Trace{
			TraceId: traceInfo.TraceId,
			SpanId:  spanId,
			UserId:  traceInfo.UserId,
			Method:  traceInfo.Method,
			Path:    req.URL.Host + req.URL.Path,
			Status:  gCtx.Writer.Status(),
			Headers: getHeader(h.header),
		}
		gCtx.Set(TraceCtx, newTrace)
		traceInfo = newTrace
		cx = context.WithValue(gCtx.Request.Context(), TraceCtx, traceInfo)
	} else {
		if h.header.Get(SpanId) != "" {
			traceInfo.SpanId = h.header.Get(SpanId)
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
			Method:  h.method,
			Path:    h.url,
			Status:  200,
			Headers: getHeader(h.header),
		}
		cx = context.WithValue(cx, TraceCtx, traceInfo)
	}

	for k, v := range traceInfo.Headers {
		if v1, ok := v.(string); ok {
			req.Header.Set(k, v1)
		}
	}
	req.Header.Set(SpanId, spanId)
	req.Header.Set(TraceId, traceInfo.TraceId)
	return cx
}

// 发起http请求,获取响应并设置对应的值
func (h *HTTPRequest) doRequest(ctx context.Context) *HTTPRespone {
	if ctx == nil {
		ctx = context.Background()
	}

	u, err := url.Parse(h.url)
	if err != nil {
		return &HTTPRespone{Error: err}
	}

	// 注入查询参数
	if len(h.params) > 0 {
		q := u.Query()
		for k, vs := range h.params {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	Infotf(ctx, "request %v: %v", h.method, u.String())

	req, err := http.NewRequestWithContext(ctx, h.method, u.String(), nil)
	if err != nil {
		return &HTTPRespone{Error: err}
	}

	// 设置 Header
	req.Header = h.header.Clone()

	h.setBody(req)
	newCtx := h.setTrace(ctx, req)
	req = req.WithContext(newCtx)
	h.setReferer(newCtx, req)

	if h.transport != nil {
		h.client.Transport = h.transport
	}
	h.client.Timeout = h.timeout

	var resp *http.Response
	for i := 0; h.retries == -1 || i <= h.retries; i++ {
		if i > 0 {
			Infotf(newCtx, "retry request %v: %v, times: %d", h.method, u.String(), i)
		}
		resp, err = h.client.Do(req)
		if err == nil {
			break
		}
		if newCtx.Err() != nil {
			return &HTTPRespone{Error: newCtx.Err()}
		}
	}

	if err != nil {
		return &HTTPRespone{Error: err}
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return &HTTPRespone{Error: err, responseCode: resp.StatusCode}
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return &HTTPRespone{Error: err, responseCode: resp.StatusCode}
	}

	hr := &HTTPRespone{
		responseCode:   resp.StatusCode,
		responseCookie: resp.Cookies(),
		responseHeader: resp.Header,
		bytes:          body,
		response:       resp,
	}
	h.response = hr
	return hr
}

// Bytes 获取响应体的字节切片
func (zr *HTTPRespone) Bytes() ([]byte, error) {
	if zr.Error != nil {
		return nil, zr.Error
	}
	return zr.bytes, nil
}

// String 获取响应体的字符串形式
func (zr *HTTPRespone) String() (string, error) {
	if zr.Error != nil {
		return "", zr.Error
	}
	return string(zr.bytes), nil
}

// Response 获取原始 http.Response
func (zr *HTTPRespone) Response() (*http.Response, error) {
	return zr.response, zr.Error
}

// ToJSON 将响应体反序列化为 JSON 对象
func (zr *HTTPRespone) ToJSON(ret any) error {
	if zr.Error != nil {
		return zr.Error
	}
	if zr.bytes == nil {
		return errors.New("body empty")
	}
	err := json.Unmarshal(zr.bytes, ret)
	return err
}
