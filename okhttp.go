package gohera

import (
	"github.com/metlive/gohera/zhttp"
	"net/http"
	"time"
)

func NewRequest() *zhttp.HTTPRequest {
	return &zhttp.HTTPRequest{
		Client: &http.Client{},
		Request: &http.Request{
			Header: make(http.Header),
		},
		Timeout: 3 * time.Second,
	}
}
