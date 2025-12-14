package gohera

// getHeader 将 http.Header (map[string][]string) 转换为 map[string]any
// 取每个 Header 的第一个值
func getHeader(headers map[string][]string) map[string]any {
	prefixHeaders := make(map[string]any)
	for k, v := range headers {
		if len(v) == 0 {
			prefixHeaders[k] = ""
		} else {
			prefixHeaders[k] = v[0]
		}
	}
	return prefixHeaders
}
