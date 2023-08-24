package gohera

import (
	"time"

	"github.com/spf13/viper"
)

var (
	config *viper.Viper
)

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

func GetConfig(key string) interface{} {
	return config.Get(key)
}

func GetDefaultString(key, defaultValue string) string {
	if value := config.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

func GetString(key string) string {
	return config.GetString(key)
}

func GetBool(key string) bool {
	return config.GetBool(key)
}

func GetInt(key string) int {
	return config.GetInt(key)
}

func GetInt32(key string) int32 {
	return config.GetInt32(key)
}

func GetInt64(key string) int64 {
	return config.GetInt64(key)
}

func GetUint(key string) uint {
	return config.GetUint(key)
}

func GetUint32(key string) uint32 {
	return config.GetUint32(key)
}

func GetUint64(key string) uint64 {
	return config.GetUint64(key)
}

func GetFloat64(key string) float64 {
	return config.GetFloat64(key)
}

func GetTime(key string) time.Time {
	return config.GetTime(key)
}

func GetDuration(key string) time.Duration {
	return config.GetDuration(key)
}

func GetStringSlice(key string) []string {
	return config.GetStringSlice(key)
}

func GetStringMap(key string) map[string]any {
	return config.GetStringMap(key)
}

func GetStringMapString(key string) map[string]string {
	return config.GetStringMapString(key)
}

func GetStringMapStringSlice(key string) map[string][]string {
	return config.GetStringMapStringSlice(key)
}

func IsSet(key string) bool {
	return config.IsSet(key)
}

func UnmarshalKey(key string, rawVal interface{}) error {
	return config.UnmarshalKey(key, rawVal)
}
