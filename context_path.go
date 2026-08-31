package gohera

import (
	"net/http"
	"net/url"
	"strings"
)

// ContextPathHandler 为 handler 增加可配置接口前缀（config key: http.context_path）。
//
// Gin 的路由匹配先于中间件执行，因此前缀剥离必须在引擎外层完成，
// 用在 http.Server 的 Handler 上，而非 engine.Use：
//
//	srv := &http.Server{Addr: addr, Handler: gohera.ContextPathHandler(engine)}
//
// 行为：
//   - 命中前缀的请求剥去前缀后再交给 handler（如 /myapp/api/v1/* → /api/v1/*），
//     业务路由按根路径注册即可，无需感知前缀
//   - 未命中的请求原样放行，根路径（本地直连、探针、本地代理）始终可用
//   - 未配置或配置为空 / "/" 时等价于直接使用原 handler
//
// 前缀逐请求读取配置快照：本地文件变更或远程配置热更新即时生效，无需重启。
func ContextPathHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := normalizeContextPath(GetString("http.context_path"))
		if prefix != "" && matchContextPath(r.URL, prefix) {
			r = withStrippedPrefix(r, prefix)
		}
		h.ServeHTTP(w, r)
	})
}

// normalizeContextPath 归一化为以 / 开头、不以 / 结尾；空值与 "/" 视为无前缀。
func normalizeContextPath(raw string) string {
	p := strings.Trim(strings.TrimSpace(raw), "/")
	if p == "" {
		return ""
	}
	return "/" + p
}

// matchContextPath 判断请求路径是否落在前缀之下（前缀本身或其子路径）。
func matchContextPath(u *url.URL, prefix string) bool {
	return u.Path == prefix || strings.HasPrefix(u.Path, prefix+"/")
}

// withStrippedPrefix 浅拷贝请求与 URL 后剥去前缀，不改动下游可见的原请求对象。
func withStrippedPrefix(r *http.Request, prefix string) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
	if r2.URL.Path == "" {
		r2.URL.Path = "/"
	}
	r2.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, prefix)
	return r2
}
