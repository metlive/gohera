package gohera

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var nacosMergeMu sync.Mutex

// mergeRemoteIntoApp 将远程/兜底配置合并进运行时 viper（顶层 key 覆盖），并刷新缓存。
// 保留原 config 实例，以免丢失 WatchConfig。
func mergeRemoteIntoApp(remoteContent, format string) error {
	remote, err := parseConfigContent(remoteContent, format)
	if err != nil {
		return err
	}

	nacosMergeMu.Lock()
	defer nacosMergeMu.Unlock()

	if err := config.MergeConfigMap(remote.AllSettings()); err != nil {
		return fmt.Errorf("merge remote config: %w", err)
	}
	return refreshCache()
}

func mergeLocalFallbackFile(path, format string) error {
	if !localFileExists(path) {
		return fmt.Errorf("local fallback config not found: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read local fallback %s: %w", path, err)
	}
	return mergeRemoteIntoApp(string(data), format)
}

func parseConfigContent(content, format string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType(normalizeConfigType(format))
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, fmt.Errorf("parse remote config: %w", err)
	}
	return v, nil
}

func normalizeConfigType(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "yml":
		return "yaml"
	case "toml", "yaml", "json", "properties":
		return strings.ToLower(strings.TrimSpace(format))
	default:
		return "yaml"
	}
}
