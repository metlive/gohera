package nacos

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

func getHTTPClient(cfg *bootstrapConfig) (config_client.IConfigClient, error) {
	nacosHTTPOnce.Do(func() {
		addr, err := parseAddr(cfg.ServerAddr, cfg.GrpcPort)
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

func fetchConfigHTTP(cfg *bootstrapConfig) (string, error) {
	client, err := getHTTPClient(cfg)
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

func startListenHTTP(s *Source, cfg *bootstrapConfig) error {
	client, err := getHTTPClient(cfg)
	if err != nil {
		return fmt.Errorf("create nacos http client: %w", err)
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
		OnChange: func(_, group, dataId, data string) {
			if err := s.merge(data, cfg.ConfigFormat); err != nil {
				fmt.Fprintf(os.Stderr, "[gohera] nacos http hot reload failed dataId=%s group=%s: %v\n", dataId, group, err)
				return
			}
			fmt.Fprintf(os.Stderr, "[gohera] nacos http config hot reloaded: dataId=%s group=%s\n", dataId, group)
		},
	})
}
