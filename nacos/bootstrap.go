package nacos

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/metlive/gohera/internal/configutil"
	"github.com/spf13/cast"
)

// bootstrapConfig 来自 bootstrap.yaml / 环境变量，描述 Nacos 连接与拉取行为。
// mode: http（nacos-sdk-go v1，兼容 OpenAPI/网关）| grpc（nacos-sdk-go v2）。
type bootstrapConfig struct {
	Enabled           bool
	Mode              string // http（默认）| grpc
	ServerAddr        string
	Namespace         string
	AccessKey         string
	SecretKey         string
	Username          string
	Password          string
	Prefix            string
	ConfigFormat      string
	FailLocalFallback bool
	DataIDByEnv       bool
	TimeoutMs         uint64
	DataID            string
	Group             string
	LocalPath         string // 拉取失败或未启用时可合并的本地兜底文件
	CacheDir          string
	LogDir            string
	GrpcPort          uint64 // grpc 模式可选；0 表示 SDK 默认 Port+1000
}

func (s *Source) loadBootstrap() (*bootstrapConfig, error) {
	m, err := loadBootstrapFile(s.SearchPaths)
	if err != nil {
		return nil, err
	}

	str := func(key string) string {
		v, _ := configutil.Lookup(m, key)
		out, _ := cast.ToStringE(v)
		return out
	}
	boolean := func(key string) bool {
		v, _ := configutil.Lookup(m, key)
		out, _ := cast.ToBoolE(v)
		return out
	}
	u64 := func(key string) uint64 {
		v, _ := configutil.Lookup(m, key)
		out, _ := cast.ToUint64E(v)
		return out
	}

	cfg := &bootstrapConfig{
		Enabled:           boolean("nacos.enabled"),
		Mode:              firstNonEmpty(str("nacos.mode"), "http"),
		ServerAddr:        firstNonEmpty(str("nacos.serveraddr"), str("nacos.server-addr")),
		Namespace:         str("nacos.namespace"),
		AccessKey:         firstNonEmpty(str("nacos.accesskey"), str("nacos.access-key")),
		SecretKey:         firstNonEmpty(str("nacos.secretkey"), str("nacos.secret-key")),
		Username:          str("nacos.username"),
		Password:          str("nacos.password"),
		Prefix:            str("nacos.prefix"),
		ConfigFormat:      firstNonEmpty(str("nacos.configformat"), str("nacos.config-format")),
		FailLocalFallback: boolean("nacos.faillocalfallback"),
		DataIDByEnv:       boolean("nacos.dataidbyenv"),
		TimeoutMs:         u64("nacos.timeoutms"),
		DataID:            firstNonEmpty(str("nacos.dataid"), str("nacos.data-id")),
		Group:             str("nacos.group"),
		LocalPath: firstNonEmpty(
			str("nacos.localpath"),
			str("nacos.local-path"),
			str("nacos.localfallbackpath"),
		),
		CacheDir: firstNonEmpty(str("nacos.cachedir"), str("nacos.cache-dir")),
		LogDir:   firstNonEmpty(str("nacos.logdir"), str("nacos.log-dir")),
		GrpcPort: firstNonZeroUint64(u64("nacos.grpcport"), u64("nacos.grpc-port")),
	}

	applyEnvOverrides(cfg)
	s.applyDefaults(cfg)
	return cfg, nil
}

// loadBootstrapFile 在 SearchPaths 中查找 bootstrap.yaml/yml/json/toml，
// 找到则加载（key 已 lowercase），找不到返回 nil（与现网 _ = ReadInConfig 一致）。
func loadBootstrapFile(searchPaths []string) (map[string]any, error) {
	exts := []string{".yaml", ".yml", ".json", ".toml"}
	for _, dir := range searchPaths {
		for _, ext := range exts {
			path := filepath.Join(dir, "bootstrap"+ext)
			if !localFileExists(path) {
				continue
			}
			m, err := configutil.LoadFile(path)
			if err != nil {
				return nil, fmt.Errorf("load bootstrap %s: %w", path, err)
			}
			return m, nil
		}
	}
	return nil, nil
}

func (s *Source) applyDefaults(cfg *bootstrapConfig) {
	if cfg.Mode == "" {
		cfg.Mode = "http"
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = 10000
	}
	if cfg.DataID == "" && cfg.Prefix != "" {
		cfg.DataID = cfg.Prefix
	}
	if cfg.DataIDByEnv && cfg.DataID != "" {
		suffix := "-" + s.currentEnv()
		if !strings.HasSuffix(cfg.DataID, suffix) {
			cfg.DataID = cfg.DataID + suffix
		}
	}
	if cfg.ConfigFormat == "" {
		cfg.ConfigFormat = inferConfigFormat(cfg.DataID)
	}
	if cfg.Group == "" {
		cfg.Group = "DEFAULT_GROUP"
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = "./.nacos/cache"
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "./.nacos/log"
	}
	if cfg.LocalPath == "" {
		// 约定：configs/nacos.{env}.yaml 作为本地兜底
		cfg.LocalPath = fmt.Sprintf("configs/nacos.%s.yaml", s.currentEnv())
	}
}

func applyEnvOverrides(cfg *bootstrapConfig) {
	if val := os.Getenv("NACOS_ENABLED"); val != "" {
		cfg.Enabled = val == "1" || strings.EqualFold(val, "true")
	}
	if val := os.Getenv("NACOS_MODE"); val != "" {
		cfg.Mode = val
	}
	if val := os.Getenv("NACOS_SERVER_ADDR"); val != "" {
		cfg.ServerAddr = val
	}
	if val := os.Getenv("NACOS_NAMESPACE"); val != "" {
		cfg.Namespace = val
	}
	if val := os.Getenv("NACOS_ACCESS_KEY"); val != "" {
		cfg.AccessKey = val
	}
	if val := os.Getenv("NACOS_SECRET_KEY"); val != "" {
		cfg.SecretKey = val
	}
	if val := os.Getenv("NACOS_USERNAME"); val != "" {
		cfg.Username = val
	}
	if val := os.Getenv("NACOS_PASSWORD"); val != "" {
		cfg.Password = val
	}
	if val := os.Getenv("NACOS_PREFIX"); val != "" {
		cfg.Prefix = val
	}
	if val := os.Getenv("NACOS_DATA_ID"); val != "" {
		cfg.DataID = val
	}
	if val := os.Getenv("NACOS_GROUP"); val != "" {
		cfg.Group = val
	}
	if val := os.Getenv("NACOS_CONFIG_FORMAT"); val != "" {
		cfg.ConfigFormat = val
	}
	if val := os.Getenv("NACOS_FAIL_LOCAL_FALLBACK"); val != "" {
		cfg.FailLocalFallback = val == "1" || strings.EqualFold(val, "true")
	}
	if val := os.Getenv("NACOS_DATA_ID_BY_ENV"); val != "" {
		cfg.DataIDByEnv = val == "1" || strings.EqualFold(val, "true")
	}
	if val := os.Getenv("NACOS_LOCAL_PATH"); val != "" {
		cfg.LocalPath = val
	}
	if val := os.Getenv("NACOS_GRPC_PORT"); val != "" {
		if p, err := strconv.ParseUint(val, 10, 64); err == nil {
			cfg.GrpcPort = p
		}
	}
}

func inferConfigFormat(dataID string) string {
	switch strings.ToLower(filepath.Ext(dataID)) {
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	default:
		return "yaml"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func firstNonZeroUint64(values ...uint64) uint64 {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func localFileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
