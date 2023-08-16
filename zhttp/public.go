package zhttp

import (
	"net/http"
	"time"
)

const (
	METHODGET  = "GET"
	METHODPOST = "POST"
	METHODPUT  = "PUT"
)

type HTTPRequest struct {
	Client    *http.Client
	Request   *http.Request
	transport http.RoundTripper
	Timeout   time.Duration
	response  *HTTPRespone
}

type HTTPRespone struct {
	responseCode   int
	responseHeader http.Header
	responseCookie []*http.Cookie
	bytes          []byte
	error          error
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

func (b *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	b.transport = transport
	return b
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
