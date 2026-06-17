package gohera

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// SSEEvent 表示单条 SSE 事件
type SSEEvent struct {
	ID    string // 事件 ID（对应 id: 字段）
	Event string // 事件类型（对应 event: 字段，为空时默认为 "message"）
	Data  string // 事件数据（多条 data: 行以 \n 拼接）
	Retry int    // 重连间隔毫秒（对应 retry: 字段，0 表示未设置）
}

// SSEHandler SSE 事件回调，返回 error 会中断流读取
type SSEHandler func(event *SSEEvent) error

// StreamConfig 流式请求配置
type StreamConfig struct {
	URL     string            // 请求地址
	Method  string            // 请求方法，默认 GET
	Headers map[string]string // 请求头
	Body    []byte            // 请求体（POST 等需要）
	Timeout time.Duration     // 连接超时，默认 30 秒
}

// Stream 发起流式请求，返回 io.ReadCloser 由调用者自行读取
// 调用者负责关闭返回的 body
func Stream(cfg *StreamConfig) (io.ReadCloser, *http.Response, error) {
	return StreamWithContext(context.Background(), cfg)
}

// StreamWithContext 带 context 的流式请求
func StreamWithContext(ctx context.Context, cfg *StreamConfig) (io.ReadCloser, *http.Response, error) {
	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	var bodyReader io.Reader
	if len(cfg.Body) > 0 {
		bodyReader = bytes.NewReader(cfg.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, cfg.URL, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp, nil
}

// StreamSSE 发起 SSE 请求，逐事件回调
func StreamSSE(cfg *StreamConfig, handler SSEHandler) error {
	return StreamSSEWithContext(context.Background(), cfg, handler)
}

// StreamSSEWithContext 带 context 的 SSE 请求，支持取消和超时控制
func StreamSSEWithContext(ctx context.Context, cfg *StreamConfig, handler SSEHandler) error {
	body, resp, err := StreamWithContext(ctx, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, body)

	reader := bufio.NewScanner(body)
	// 增大缓冲区以支持较长的 data 行
	reader.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var event SSEEvent
	var dataLines []string

	for reader.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := reader.Text()

		// 空行 = 事件边界
		if line == "" {
			if len(dataLines) > 0 {
				event.Data = strings.Join(dataLines, "\n")
				if err := handler(&event); err != nil {
					return err
				}
				event = SSEEvent{}
				dataLines = dataLines[:0]
			}
			continue
		}

		// 注释行，跳过
		if strings.HasPrefix(line, ":") {
			continue
		}

		// 解析 field:value
		colon := strings.IndexByte(line, ':')
		if colon == -1 {
			continue
		}

		field := line[:colon]
		value := line[colon+1:]
		// 去掉冒号后的一个空格（SSE 规范）
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		switch field {
		case "id":
			event.ID = value
		case "event":
			event.Event = value
		case "data":
			dataLines = append(dataLines, value)
		case "retry":
			// 已简化，忽略解析失败的 retry 值
		}
	}

	// 连接关闭时，发送最后一个未完成的事件
	if len(dataLines) > 0 {
		event.Data = strings.Join(dataLines, "\n")
		if err := handler(&event); err != nil {
			return err
		}
	}

	return reader.Err()
}
