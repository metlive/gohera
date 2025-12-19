package gohera

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var config = viper.New()
var configLoaded bool
var configCache atomic.Pointer[map[string]any]

func init() {
	// 尝试在包加载时初始化配置，以便包级别的变量初始化可以获取到配置
	_ = initAppConfig()
}

// initAppConfig 初始化应用配置
// 按照优先级从当前目录、./config、./configs 加载 app.toml/yaml/json 等配置文件
func initAppConfig() error {
	if configLoaded {
		return nil
	}

	config.SetConfigName("app")
	config.AddConfigPath("./")
	config.AddConfigPath("./config")
	config.AddConfigPath("./configs")

	err := config.ReadInConfig()
	if err != nil {
		return err
	}

	config.SetEnvPrefix("APP")
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		_ = refreshCache()
	})
	config.WatchConfig()

	_ = refreshCache()
	configLoaded = true
	return nil
}

// refreshCache 刷新配置缓存，将所有配置扁平化存入 atomic.Pointer
func refreshCache() error {
	allSettings := config.AllSettings()
	flatCache := make(map[string]any)
	flattenSettings("", allSettings, flatCache)
	configCache.Store(&flatCache)
	return nil
}

// flattenSettings 递归扁平化配置项
func flattenSettings(prefix string, settings map[string]any, cache map[string]any) {
	for k, v := range settings {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		cache[fullKey] = v
		if subMap, ok := v.(map[string]any); ok {
			flattenSettings(fullKey, subMap, cache)
		}
	}
}

// GetConfig 获取原始配置值
func GetConfig(key string) any {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return v
		}
	}
	return config.Get(key)
}

// GetDefaultString 获取字符串配置，如果不存在则返回默认值
func GetDefaultString(key, defaultValue string) string {
	if value := GetString(key); value != "" {
		return value
	}
	return defaultValue
}

// GetString 获取字符串配置
func GetString(key string) string {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToString(v)
		}
	}
	return config.GetString(key)
}

// GetBool 获取布尔值配置
func GetBool(key string) bool {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToBool(v)
		}
	}
	return config.GetBool(key)
}

// GetInt 获取 int 配置
func GetInt(key string) int {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToInt(v)
		}
	}
	return config.GetInt(key)
}

// GetInt32 获取 int32 配置
func GetInt32(key string) int32 {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToInt32(v)
		}
	}
	return config.GetInt32(key)
}

// GetInt64 获取 int64 配置
func GetInt64(key string) int64 {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToInt64(v)
		}
	}
	return config.GetInt64(key)
}

// GetUint 获取 uint 配置
func GetUint(key string) uint {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToUint(v)
		}
	}
	return config.GetUint(key)
}

// GetUint32 获取 uint32 配置
func GetUint32(key string) uint32 {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToUint32(v)
		}
	}
	return config.GetUint32(key)
}

// GetUint64 获取 uint64 配置
func GetUint64(key string) uint64 {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToUint64(v)
		}
	}
	return config.GetUint64(key)
}

// GetFloat64 获取 float64 配置
func GetFloat64(key string) float64 {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToFloat64(v)
		}
	}
	return config.GetFloat64(key)
}

// GetTime 获取时间配置
func GetTime(key string) time.Time {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToTime(v)
		}
	}
	return config.GetTime(key)
}

// GetDuration 获取时间间隔配置
func GetDuration(key string) time.Duration {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToDuration(v)
		}
	}
	return config.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置
func GetStringSlice(key string) []string {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToStringSlice(v)
		}
	}
	return config.GetStringSlice(key)
}

// GetStringMap 获取 map[string]any 配置
func GetStringMap(key string) map[string]any {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToStringMap(v)
		}
	}
	return config.GetStringMap(key)
}

// GetStringMapString 获取 map[string]string 配置
func GetStringMapString(key string) map[string]string {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToStringMapString(v)
		}
	}
	return config.GetStringMapString(key)
}

// GetStringMapStringSlice 获取 map[string][]string 配置
func GetStringMapStringSlice(key string) map[string][]string {
	if cache := configCache.Load(); cache != nil {
		if v, ok := (*cache)[key]; ok {
			return cast.ToStringMapStringSlice(v)
		}
	}
	return config.GetStringMapStringSlice(key)
}

// IsSet 检查配置项是否存在
func IsSet(key string) bool {
	if cache := configCache.Load(); cache != nil {
		if _, ok := (*cache)[key]; ok {
			return true
		}
	}
	return config.IsSet(key)
}

// UnmarshalKey 将配置反序列化到结构体
func UnmarshalKey(key string, rawVal any) error {
	rv := reflect.ValueOf(rawVal)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("rawVal must be a non-nil pointer")
	}
	return config.UnmarshalKey(key, rawVal)
}
