package nacos

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/metlive/gohera/internal/configutil"
)

// merge 将远程/兜底配置解析后经 Merge 回调写回应用配置（深合并）并刷新缓存。
func (s *Source) merge(remoteContent, format string) error {
	if strings.TrimSpace(remoteContent) == "" {
		return errors.New("nacos config is empty")
	}
	remote, err := parseConfigContent(remoteContent, format)
	if err != nil {
		return err
	}
	return s.Merge(remote)
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

func parseConfigContent(content, format string) (map[string]any, error) {
	t, err := normalizeConfigType(format)
	if err != nil {
		return nil, err
	}
	return configutil.Parse([]byte(content), t)
}

func normalizeConfigType(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "yml":
		return "yaml", nil
	case "yaml", "toml", "json":
		return strings.ToLower(strings.TrimSpace(format)), nil
	case "properties":
		return "", fmt.Errorf("properties format is not supported (use yaml/toml/json)")
	default:
		return "yaml", nil
	}
}
