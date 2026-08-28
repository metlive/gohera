package gohera

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	nacosGRPCOnce   sync.Once
	nacosGRPCClient config_client.IConfigClient
	nacosGRPCErr    error
)

func getNacosGRPCClient(cfg *nacosBootstrap) (config_client.IConfigClient, error) {
	nacosGRPCOnce.Do(func() {
		addr, err := parseNacosAddr(cfg.ServerAddr, cfg.GrpcPort)
		if err != nil {
			nacosGRPCErr = err
			return
		}

		cc := *constant.NewClientConfig(
			constant.WithNamespaceId(cfg.Namespace),
			constant.WithTimeoutMs(cfg.TimeoutMs),
			constant.WithNotLoadCacheAtStart(true),
			constant.WithLogDir(cfg.LogDir),
			constant.WithCacheDir(cfg.CacheDir),
			constant.WithLogLevel("warn"),
			constant.WithAccessKey(cfg.AccessKey),
			constant.WithSecretKey(cfg.SecretKey),
			constant.WithUsername(cfg.Username),
			constant.WithPassword(cfg.Password),
		)

		scOpts := []constant.ServerOption{
			constant.WithScheme(addr.Scheme),
			constant.WithContextPath(addr.ContextPath),
		}
		if addr.GrpcPort > 0 {
			scOpts = append(scOpts, constant.WithGrpcPort(addr.GrpcPort))
		}
		sc := []constant.ServerConfig{
			*constant.NewServerConfig(addr.IpAddr, addr.Port, scOpts...),
		}

		nacosGRPCClient, nacosGRPCErr = clients.NewConfigClient(vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		})
	})
	return nacosGRPCClient, nacosGRPCErr
}

func fetchNacosConfigGRPC(cfg *nacosBootstrap) (string, error) {
	client, err := getNacosGRPCClient(cfg)
	if err != nil {
		return "", fmt.Errorf("create nacos grpc client: %w", err)
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

func startNacosListenGRPC(cfg *nacosBootstrap) error {
	client, err := getNacosGRPCClient(cfg)
	if err != nil {
		return fmt.Errorf("create nacos grpc client: %w", err)
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
		OnChange: func(_, group, dataId, data string) {
			if err := mergeRemoteIntoApp(data, cfg.ConfigFormat); err != nil {
				fmt.Fprintf(os.Stderr, "[gohera] nacos grpc hot reload failed dataId=%s group=%s: %v\n", dataId, group, err)
				return
			}
			fmt.Fprintf(os.Stderr, "[gohera] nacos grpc config hot reloaded: dataId=%s group=%s\n", dataId, group)
		},
	})
}
