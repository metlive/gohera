package nacos

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var nacosMergeMu sync.Mutex

// merge 将远程/兜底配置解析后经 Merge 回调写回应用配置（顶层 key 覆盖）并刷新缓存。
func (s *Source) merge(remoteContent, format string) error {
	remote, err := parseConfigContent(remoteContent, format)
	if err != nil {
		return err
	}

	nacosMergeMu.Lock()
	defer nacosMergeMu.Unlock()

	return s.Merge(remote.AllSettings())
}

func (s *Source) mergeLocalFallback(path, format string) error {
	if !localFileExists(path) {
		return fmt.Errorf("local fallback config not found: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read local fallback %s: %w", path, err)
	}
	return s.merge(string(data), format)
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
