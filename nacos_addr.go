package gohera

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// nacosAddr 解析后的 Nacos 服务地址（HTTP / gRPC 共用）。
type nacosAddr struct {
	Scheme      string
	IpAddr      string
	Port        uint64
	ContextPath string
	GrpcPort    uint64 // 0 表示交由 SDK 按 Port+1000 推导
}

func parseNacosAddr(serverAddr string, grpcPort uint64) (*nacosAddr, error) {
	addr := strings.TrimSpace(serverAddr)
	if addr == "" {
		return nil, fmt.Errorf("nacos.serverAddr is empty")
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("parse nacos.serverAddr %q: %w", serverAddr, err)
	}

	scheme := u.Scheme
	port := u.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	portNum, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse nacos port %q: %w", port, err)
	}

	contextPath := "/nacos"
	if path := strings.TrimSuffix(u.Path, "/"); path != "" {
		if strings.HasSuffix(path, "/nacos") {
			contextPath = path
		} else {
			contextPath = path + "/nacos"
		}
	}

	return &nacosAddr{
		Scheme:      scheme,
		IpAddr:      u.Hostname(),
		Port:        portNum,
		ContextPath: contextPath,
		GrpcPort:    grpcPort,
	}, nil
}
