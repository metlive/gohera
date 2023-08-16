package zhttp

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	FORMCONTENTTYPE = "application/x-www-form-urlencoded"
	JSONCONTENTTYPE = "application/json"
)

/*
*

	内部接口请求请使用带ctx的方法(带链路追踪),请求外部接口用不带ctx的
*/
func (h *HTTPRequest) GetCtx(ctx *gin.Context, reqUrl string) *HTTPRespone {
	h.Request.Header.Add("Content-Type", FORMCONTENTTYPE)
	resp := h.setTrace(ctx).setNextRpcId(ctx).addReferer(ctx).doRequest(METHODGET, reqUrl)
	return resp
}

func (h *HTTPRequest) PostCtx(ctx *gin.Context, reqUrl string, params map[string]string) *HTTPRespone {
	args := &url.Values{}
	for key, value := range params {
		args.Add(key, value)
	}

	h.Request.Header.Set("Content-Type", FORMCONTENTTYPE)

	resp := h.setTrace(ctx).setNextRpcId(ctx).addReferer(ctx).setBody([]byte(args.Encode())).doRequest(METHODPOST, reqUrl)

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

	resp = h.setTrace(ctx).setNextRpcId(ctx).addReferer(ctx).setBody(requestBody).doRequest(METHODPOST, reqUrl)

	return resp
}

func (h *HTTPRequest) Get(reqUrl string) *HTTPRespone {
	h.Request.Header.Add("Content-Type", FORMCONTENTTYPE)
	resp := h.doRequest(METHODGET, reqUrl)
	return resp
}

func (h *HTTPRequest) Post(reqUrl string, params map[string]string) *HTTPRespone {
	args := &url.Values{}
	for key, value := range params {
		args.Add(key, value)
	}

	h.Request.Header.Set("Content-Type", FORMCONTENTTYPE)

	resp := h.setBody([]byte(args.Encode())).doRequest(METHODPOST, reqUrl)

	return resp
}

func (h *HTTPRequest) JsonPost(reqUrl string, params interface{}) *HTTPRespone {
	h.Request.Header.Set("Content-Type", JSONCONTENTTYPE)

	requestBody, err := json.Marshal(params)
	resp := &HTTPRespone{}
	if err != nil {
		resp.error = errors.New("json marshal fail")
		return resp
	}

	resp = h.setBody(requestBody).doRequest(METHODPOST, reqUrl)
	return resp
}
