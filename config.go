package gohera

import (
	"time"

	"github.com/spf13/viper"
)

var config *viper.Viper

// initAppConfig 初始化应用配置
// 使用 viper 加载 ./config/app.toml 配置文件
func initAppConfig() error {
	if config != nil {
		return nil
	}

	config = viper.New()
	config.SetConfigName("app")
	config.AddConfigPath("./config")
	config.SetConfigType("toml")

	err := config.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// GetConfig 获取原始配置值
func GetConfig(key string) any {
	return config.Get(key)
}

// GetDefaultString 获取字符串配置，如果不存在则返回默认值
func GetDefaultString(key, defaultValue string) string {
	if value := config.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

// GetString 获取字符串配置
func GetString(key string) string {
	return config.GetString(key)
}

// GetBool 获取布尔值配置
func GetBool(key string) bool {
	return config.GetBool(key)
}

// GetInt 获取 int 配置
func GetInt(key string) int {
	return config.GetInt(key)
}

// GetInt32 获取 int32 配置
func GetInt32(key string) int32 {
	return config.GetInt32(key)
}

// GetInt64 获取 int64 配置
func GetInt64(key string) int64 {
	return config.GetInt64(key)
}

// GetUint 获取 uint 配置
func GetUint(key string) uint {
	return config.GetUint(key)
}

// GetUint32 获取 uint32 配置
func GetUint32(key string) uint32 {
	return config.GetUint32(key)
}

// GetUint64 获取 uint64 配置
func GetUint64(key string) uint64 {
	return config.GetUint64(key)
}

// GetFloat64 获取 float64 配置
func GetFloat64(key string) float64 {
	return config.GetFloat64(key)
}

// GetTime 获取时间配置
func GetTime(key string) time.Time {
	return config.GetTime(key)
}

// GetDuration 获取时间间隔配置
func GetDuration(key string) time.Duration {
	return config.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置
func GetStringSlice(key string) []string {
	return config.GetStringSlice(key)
}

// GetStringMap 获取 map[string]any 配置
func GetStringMap(key string) map[string]any {
	return config.GetStringMap(key)
}

// GetStringMapString 获取 map[string]string 配置
func GetStringMapString(key string) map[string]string {
	return config.GetStringMapString(key)
}

// GetStringMapStringSlice 获取 map[string][]string 配置
func GetStringMapStringSlice(key string) map[string][]string {
	return config.GetStringMapStringSlice(key)
}

// IsSet 检查配置项是否存在
func IsSet(key string) bool {
	return config.IsSet(key)
}

// UnmarshalKey 将配置反序列化到结构体
func UnmarshalKey(key string, rawVal any) error {
	return config.UnmarshalKey(key, rawVal)
}
