package gohera

import (
	"net/http"
	"time"
)

type GRequest struct {
	client    *http.Client
	request   *http.Request
	transport http.RoundTripper
	timeout   time.Duration
	response  *GRespone
}

type GRespone struct {
	responseCode   int
	responseHeader http.Header
	responseCookie []*http.Cookie
	bytes          []byte
	error          error
}
