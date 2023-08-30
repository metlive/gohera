package gohera

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
