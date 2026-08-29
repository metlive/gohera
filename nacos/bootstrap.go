package nacos

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
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
	v := viper.New()
	v.SetConfigName("bootstrap")
	for _, dir := range s.SearchPaths {
		v.AddConfigPath(dir)
	}
	_ = v.ReadInConfig()

	cfg := &bootstrapConfig{
		Enabled:           v.GetBool("nacos.enabled"),
		Mode:              firstNonEmpty(v.GetString("nacos.mode"), "http"),
		ServerAddr:        firstNonEmpty(v.GetString("nacos.serverAddr"), v.GetString("nacos.server-addr")),
		Namespace:         v.GetString("nacos.namespace"),
		AccessKey:         firstNonEmpty(v.GetString("nacos.accessKey"), v.GetString("nacos.access-key")),
		SecretKey:         firstNonEmpty(v.GetString("nacos.secretKey"), v.GetString("nacos.secret-key")),
		Username:          v.GetString("nacos.username"),
		Password:          v.GetString("nacos.password"),
		Prefix:            v.GetString("nacos.prefix"),
		ConfigFormat:      firstNonEmpty(v.GetString("nacos.configFormat"), v.GetString("nacos.config-format")),
		FailLocalFallback: v.GetBool("nacos.failLocalFallback"),
		DataIDByEnv:       v.GetBool("nacos.dataIdByEnv"),
		TimeoutMs:         v.GetUint64("nacos.timeoutMs"),
		DataID:            firstNonEmpty(v.GetString("nacos.dataId"), v.GetString("nacos.data-id")),
		Group:             v.GetString("nacos.group"),
		LocalPath: firstNonEmpty(
			v.GetString("nacos.localPath"),
			v.GetString("nacos.local-path"),
			v.GetString("nacos.localFallbackPath"),
		),
		CacheDir: firstNonEmpty(v.GetString("nacos.cacheDir"), v.GetString("nacos.cache-dir")),
		LogDir:   firstNonEmpty(v.GetString("nacos.logDir"), v.GetString("nacos.log-dir")),
		GrpcPort: firstNonZeroUint64(v.GetUint64("nacos.grpcPort"), v.GetUint64("nacos.grpc-port")),
	}

	applyEnvOverrides(cfg)
	s.applyDefaults(cfg)
	return cfg, nil
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
