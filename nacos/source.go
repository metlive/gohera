package nacos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source 负责加载 Nacos bootstrap，并在启用时按 mode（http|grpc）拉取远程配置、
// 经 Merge 回调写回应用配置并注册热更新监听。
type Source struct {
	DefaultEnv string // 当前环境兜底值（根传 gohera.DeployEnvDev）
	Env        func() string
	// SearchPaths bootstrap.yaml 搜索目录（根传 configSearchPaths）
	SearchPaths []string
	// Merge 将远程/兜底配置写回根 store 的 overlay 层（覆盖 base，含热更新）
	Merge func(map[string]any) error
	// MergeBase 将引导文件（bootstrap*.yaml）中 nacos 段之外的内容合入 base 层：
	// 环境级本地配置，覆盖 app.yaml 同名键、低于远程/兜底 overlay。可选，未设置时忽略。
	MergeBase func(map[string]any) error
}

// Init 在 InitApp 中、MySQL/Redis 初始化之前调用：
// 0. 引导文件的非 nacos 段（如有）经 MergeBase 合入 base（环境级本地配置）
// 1. nacos.enabled=true → 按 mode(http|grpc) 拉取并合并，注册监听
// 2. 拉取失败且 failLocalFallback → 本地值（已合入 base）继续生效；
//    显式配置的 localPath（nacos.localPath / NACOS_LOCAL_PATH）存在时合并
// 3. nacos 未启用 → 同上，仅显式 localPath 存在时合并
func (s *Source) Init() error {
	cfg, body, err := s.loadBootstrap()
	if err != nil {
		return err
	}

	if len(body) > 0 {
		if s.MergeBase == nil {
			fmt.Fprintf(os.Stderr, "[gohera] nacos bootstrap local config ignored (MergeBase not set)\n")
		} else if err := s.MergeBase(body); err != nil {
			return fmt.Errorf("merge bootstrap local config: %w", err)
		} else {
			fmt.Fprintf(os.Stderr, "[gohera] nacos bootstrap local config merged (%d sections)\n", len(body))
		}
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
			if localFileExists(cfg.LocalPath) {
				fmt.Fprintf(os.Stderr, "[gohera] nacos fetch failed (mode=%s), fallback to %s: %v\n", mode, cfg.LocalPath, err)
				return s.mergeLocalFallback(cfg.LocalPath, cfg.ConfigFormat)
			}
			// 本地值（引导文件非 nacos 段，已合入 base）继续生效
			fmt.Fprintf(os.Stderr, "[gohera] nacos fetch failed (mode=%s), running on local config: %v\n", mode, err)
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

// BootstrapExists 报告 SearchPaths 中是否存在任一 bootstrap 文件
// （bootstrap.{ext} 公共基础或 bootstrap-{env}.{ext} 当前环境覆盖）。
// 供 InitApp 校验「app 配置与 Nacos 引导至少存在其一」。
func (s *Source) BootstrapExists() bool {
	env := s.currentEnv()
	exts := []string{".yaml", ".yml", ".json", ".toml"}
	for _, dir := range s.SearchPaths {
		for _, ext := range exts {
			if localFileExists(filepath.Join(dir, "bootstrap"+ext)) ||
				localFileExists(filepath.Join(dir, "bootstrap-"+env+ext)) {
				return true
			}
		}
	}
	return false
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
