// Package configutil 提供配置内核的共享工具：koanf parser 选择与 key 归一化。
package configutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
)

// ParserForType 返回归一化类型（yaml|yml|json|toml）对应的 koanf parser。
// properties 现网 viper v1.21 已无 decoder，这里显式拒绝。
func ParserForType(t string) (koanf.Parser, error) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "yaml", "yml":
		return yaml.Parser(), nil
	case "json":
		return json.Parser(), nil
	case "toml":
		return toml.Parser(), nil
	case "properties":
		return nil, fmt.Errorf("properties format is not supported (use yaml/toml/json)")
	default:
		return nil, fmt.Errorf("unsupported config format %q", t)
	}
}

// ParserForExt 返回文件扩展名（可带点）对应的 koanf parser。
func ParserForExt(ext string) (koanf.Parser, error) {
	return ParserForType(strings.TrimPrefix(strings.ToLower(ext), "."))
}

// LoadFile 读取并解析配置文件，返回 key 已 lowercase 的嵌套 map。
func LoadFile(path string) (map[string]any, error) {
	parser, err := ParserForExt(filepath.Ext(path))
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	m, err := parser.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return Lowercase(m), nil
}

// Parse 解析原始配置内容，返回 key 已 lowercase 的嵌套 map。
func Parse(content []byte, format string) (map[string]any, error) {
	parser, err := ParserForType(format)
	if err != nil {
		return nil, err
	}
	m, err := parser.Unmarshal(content)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return Lowercase(m), nil
}

// Lowercase 递归 lowercase 嵌套 map 的全部 key，返回新 map。
func Lowercase(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = lowerVal(v)
	}
	return out
}

func lowerVal(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return Lowercase(t)
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = lowerVal(t[i])
		}
		return out
	default:
		return v
	}
}

// Lookup 在嵌套 map 中按点号路径取值。
func Lookup(m map[string]any, key string) (any, bool) {
	if key == "" {
		return m, true
	}
	var cur any = m
	for _, p := range strings.Split(key, ".") {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = mm[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
