package gohera

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	nacosHTTPOnce   sync.Once
	nacosHTTPClient config_client.IConfigClient
	nacosHTTPErr    error
)

// initNacosConfig 在 InitApp 中、MySQL/Redis 初始化之前调用：
// 1. nacos.enabled=true → 按 mode(http|grpc) 拉取并合并，注册监听
// 2. 拉取失败且 failLocalFallback → 合并 localPath
// 3. nacos 未启用 → 若存在 localPath 则合并本地兜底
func initNacosConfig() error {
	cfg, err := loadNacosBootstrap()
	if err != nil {
		return err
	}

	if !cfg.Enabled {
		if localFileExists(cfg.LocalPath) {
			if err := mergeLocalFallbackFile(cfg.LocalPath, cfg.ConfigFormat); err != nil {
				return fmt.Errorf("merge local nacos fallback: %w", err)
			}
			fmt.Fprintf(os.Stderr, "[gohera] nacos disabled, merged local fallback %s\n", cfg.LocalPath)
		}
		return nil
	}

	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = "http"
	}

	var content string
	switch mode {
	case "http":
		content, err = fetchNacosConfigHTTP(cfg)
	case "grpc":
		content, err = fetchNacosConfigGRPC(cfg)
	default:
		return fmt.Errorf("nacos.mode %q not supported (use http or grpc)", cfg.Mode)
	}

	if err != nil {
		if cfg.FailLocalFallback {
			fmt.Fprintf(os.Stderr, "[gohera] nacos fetch failed (mode=%s), fallback to %s: %v\n", mode, cfg.LocalPath, err)
			if localFileExists(cfg.LocalPath) {
				return mergeLocalFallbackFile(cfg.LocalPath, cfg.ConfigFormat)
			}
			return nil
		}
		return err
	}

	if err := mergeRemoteIntoApp(content, cfg.ConfigFormat); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[gohera] nacos config merged: mode=%s dataId=%s group=%s\n", mode, cfg.DataID, cfg.Group)

	switch mode {
	case "http":
		return startNacosListenHTTP(cfg)
	case "grpc":
		return startNacosListenGRPC(cfg)
	default:
		return nil
	}
}

func getNacosHTTPClient(cfg *nacosBootstrap) (config_client.IConfigClient, error) {
	nacosHTTPOnce.Do(func() {
		addr, err := parseNacosAddr(cfg.ServerAddr, cfg.GrpcPort)
		if err != nil {
			nacosHTTPErr = err
			return
		}
		clientConfig := constant.ClientConfig{
			NamespaceId:         cfg.Namespace,
			TimeoutMs:           cfg.TimeoutMs,
			AccessKey:           cfg.AccessKey,
			SecretKey:           cfg.SecretKey,
			Username:            cfg.Username,
			Password:            cfg.Password,
			CacheDir:            cfg.CacheDir,
			LogDir:              cfg.LogDir,
			NotLoadCacheAtStart: true,
			LogLevel:            "warn",
		}
		nacosHTTPClient, nacosHTTPErr = clients.NewConfigClient(vo.NacosClientParam{
			ClientConfig: &clientConfig,
			ServerConfigs: []constant.ServerConfig{{
				Scheme:      addr.Scheme,
				IpAddr:      addr.IpAddr,
				Port:        addr.Port,
				ContextPath: addr.ContextPath,
			}},
		})
	})
	return nacosHTTPClient, nacosHTTPErr
}

func fetchNacosConfigHTTP(cfg *nacosBootstrap) (string, error) {
	client, err := getNacosHTTPClient(cfg)
	if err != nil {
		return "", fmt.Errorf("create nacos http client: %w", err)
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
	if err != nil {
		return "", fmt.Errorf("get config dataId=%s group=%s: %w", cfg.DataID, cfg.Group, err)
	}
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("nacos config is empty")
	}
	return content, nil
}

func startNacosListenHTTP(cfg *nacosBootstrap) error {
	client, err := getNacosHTTPClient(cfg)
	if err != nil {
		return fmt.Errorf("create nacos http client: %w", err)
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
		OnChange: func(_, group, dataId, data string) {
			if err := mergeRemoteIntoApp(data, cfg.ConfigFormat); err != nil {
				fmt.Fprintf(os.Stderr, "[gohera] nacos http hot reload failed dataId=%s group=%s: %v\n", dataId, group, err)
				return
			}
			fmt.Fprintf(os.Stderr, "[gohera] nacos http config hot reloaded: dataId=%s group=%s\n", dataId, group)
		},
	})
}
