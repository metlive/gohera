package nacos

import (
	"fmt"
	"os"
	"strings"
)

// Source 负责加载 Nacos bootstrap，并在启用时按 mode（http|grpc）拉取远程配置、
// 经 Merge 回调写回应用配置并注册热更新监听。
type Source struct {
	DefaultEnv  string                     // 当前环境兜底值（根传 gohera.DeployEnvDev）
	Env         func() string              // 当前运行环境（根传 GetEnv）
	SearchPaths []string                   // bootstrap.yaml 搜索目录（根传 configSearchPaths）
	Merge       func(map[string]any) error // 将远程/兜底配置写回根 store 并刷新缓存
}

// Init 在 InitApp 中、MySQL/Redis 初始化之前调用：
// 1. nacos.enabled=true → 按 mode(http|grpc) 拉取并合并，注册监听
// 2. 拉取失败且 failLocalFallback → 合并 localPath
// 3. nacos 未启用 → 若存在 localPath 则合并本地兜底
func (s *Source) Init() error {
	cfg, err := s.loadBootstrap()
	if err != nil {
		return err
	}

	if !cfg.Enabled {
		if localFileExists(cfg.LocalPath) {
			if err := s.mergeLocalFallback(cfg.LocalPath, cfg.ConfigFormat); err != nil {
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
		content, err = fetchConfigHTTP(cfg)
	case "grpc":
		content, err = fetchConfigGRPC(cfg)
	default:
		return fmt.Errorf("nacos.mode %q not supported (use http or grpc)", cfg.Mode)
	}

	if err != nil {
		if cfg.FailLocalFallback {
			fmt.Fprintf(os.Stderr, "[gohera] nacos fetch failed (mode=%s), fallback to %s: %v\n", mode, cfg.LocalPath, err)
			if localFileExists(cfg.LocalPath) {
				return s.mergeLocalFallback(cfg.LocalPath, cfg.ConfigFormat)
			}
			return nil
		}
		return err
	}

	if err := s.merge(content, cfg.ConfigFormat); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[gohera] nacos config merged: mode=%s dataId=%s group=%s\n", mode, cfg.DataID, cfg.Group)

	switch mode {
	case "http":
		return startListenHTTP(s, cfg)
	case "grpc":
		return startListenGRPC(s, cfg)
	default:
		return nil
	}
}

// currentEnv 返回 Env() 结果，为空时回落 DefaultEnv，再回落 "dev"（保持原行为）。
func (s *Source) currentEnv() string {
	if s.Env != nil {
		if e := s.Env(); e != "" {
			return e
		}
	}
	if s.DefaultEnv != "" {
		return s.DefaultEnv
	}
	return "dev"
}
