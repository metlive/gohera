package gohera

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/metlive/gohera/internal/configutil"
	"github.com/spf13/cast"
)

var (
	configSearchPaths = []string{"./", "./config", "./configs"}
	configExtensions  = []string{".toml", ".yaml", ".yml", ".json"}
	defaultConfigName = "app"
)

var initConfigOnce sync.Once
var initConfigErr error

func init() {
	// 尝试在包加载时初始化配置，以便包级别的变量初始化可以获取到配置
	if err := initAppConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "[gohera] config init warning: %v\n", err)
	}
}

// initAppConfig 初始化应用配置
// 扫描 ./、./config、./configs 目录，按优先级发现并加载配置文件
func initAppConfig() error {
	initConfigOnce.Do(func() {
		path, err := discoverConfigFile()
		if err != nil {
			initConfigErr = err
			return
		}
		initConfigErr = store.init(path)
	})
	return initConfigErr
}

// discoverConfigFile 按优先级发现配置文件：
// 1. APP_CONFIG_FILE 指定绝对路径
// 2. APP_CONFIG 或默认 app 作为文件名，在搜索路径中按扩展名匹配
// 3. 某目录下仅有一个配置文件时使用该文件
func discoverConfigFile() (string, error) {
	if path := os.Getenv("APP_CONFIG_FILE"); path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("APP_CONFIG_FILE %q: %w", path, err)
		}
		return path, nil
	}

	preferredName := os.Getenv("APP_CONFIG")
	if preferredName == "" {
		preferredName = defaultConfigName
	}

	for _, dir := range configSearchPaths {
		if path, ok := findConfigByName(dir, preferredName); ok {
			return path, nil
		}
	}

	if os.Getenv("APP_CONFIG") != "" {
		return "", fmt.Errorf("config file %q not found in %v", preferredName, configSearchPaths)
	}

	for _, dir := range configSearchPaths {
		candidates, err := listConfigFiles(dir)
		if err != nil {
			return "", err
		}
		switch len(candidates) {
		case 0:
			continue
		case 1:
			return candidates[0], nil
		default:
			return "", fmt.Errorf("multiple config files in %s: %v", dir, candidates)
		}
	}

	return "", fmt.Errorf("no config file found in %v", configSearchPaths)
}

func findConfigByName(dir, name string) (string, bool) {
	for _, ext := range configExtensions {
		path := filepath.Join(dir, name+ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func listConfigFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() || !isConfigFile(entry.Name()) {
			continue
		}
		candidates = append(candidates, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(candidates)
	return candidates, nil
}

func isConfigFile(name string) bool {
	for _, ext := range configExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// GetConfig 获取原始配置值
func GetConfig(key string) any {
	v, _ := store.lookup(key)
	return deepCopyValue(v)
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
	v, _ := store.lookup(key)
	return cast.ToString(v)
}

// GetBool 获取布尔值配置
func GetBool(key string) bool {
	v, _ := store.lookup(key)
	return cast.ToBool(v)
}

// GetInt 获取 int 配置
func GetInt(key string) int {
	v, _ := store.lookup(key)
	return cast.ToInt(v)
}

// GetInt32 获取 int32 配置
func GetInt32(key string) int32 {
	v, _ := store.lookup(key)
	return cast.ToInt32(v)
}

// GetInt64 获取 int64 配置
func GetInt64(key string) int64 {
	v, _ := store.lookup(key)
	return cast.ToInt64(v)
}

// GetUint 获取 uint 配置
func GetUint(key string) uint {
	v, _ := store.lookup(key)
	return cast.ToUint(v)
}

// GetUint32 获取 uint32 配置
func GetUint32(key string) uint32 {
	v, _ := store.lookup(key)
	return cast.ToUint32(v)
}

// GetUint64 获取 uint64 配置
func GetUint64(key string) uint64 {
	v, _ := store.lookup(key)
	return cast.ToUint64(v)
}

// GetFloat64 获取 float64 配置
func GetFloat64(key string) float64 {
	v, _ := store.lookup(key)
	return cast.ToFloat64(v)
}

// GetTime 获取时间配置
func GetTime(key string) time.Time {
	v, _ := store.lookup(key)
	return cast.ToTime(v)
}

// GetDuration 获取时间间隔配置
func GetDuration(key string) time.Duration {
	v, _ := store.lookup(key)
	return cast.ToDuration(v)
}

// GetStringSlice 获取字符串切片配置
func GetStringSlice(key string) []string {
	v, _ := store.lookup(key)
	return cloneStringSlice(cast.ToStringSlice(deepCopyValue(v)))
}

func cloneStringSlice(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

// GetStringMap 获取 map[string]any 配置
func GetStringMap(key string) map[string]any {
	v, _ := store.lookup(key)
	return cast.ToStringMap(deepCopyValue(v))
}

// GetStringMapString 获取 map[string]string 配置
func GetStringMapString(key string) map[string]string {
	v, _ := store.lookup(key)
	return cast.ToStringMapString(deepCopyValue(v))
}

// GetStringMapStringSlice 获取 map[string][]string 配置
func GetStringMapStringSlice(key string) map[string][]string {
	v, _ := store.lookup(key)
	return cast.ToStringMapStringSlice(deepCopyValue(v))
}

// IsSet 检查配置项是否存在
func IsSet(key string) bool {
	_, ok := store.lookup(key)
	return ok
}

// Unmarshal 将整份已合并配置（文件 + Nacos + env）解码到 rawVal。
// rawVal 必须是非空指针。
func Unmarshal(rawVal any) error {
	if err := validatePtr(rawVal); err != nil {
		return err
	}
	snap := store.snapshot.Load()
	if snap == nil {
		return errors.New("config: not loaded")
	}
	return decode(snap.nested, rawVal)
}

// UnmarshalKey 将配置子树反序列化到结构体
func UnmarshalKey(key string, rawVal any) error {
	if err := validatePtr(rawVal); err != nil {
		return err
	}
	snap := store.snapshot.Load()
	if snap == nil {
		return errors.New("config: not loaded")
	}
	v, ok := configutil.Lookup(snap.nested, strings.ToLower(key))
	if !ok {
		return fmt.Errorf("config key %q not found", key)
	}
	return decode(v, rawVal)
}

func validatePtr(rawVal any) error {
	rv := reflect.ValueOf(rawVal)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("rawVal must be a non-nil pointer")
	}
	return nil
}

// decode 使用与现网 viper 完全同构的 mapstructure 配置解码。
func decode(input, output any) error {
	cfg := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToWeakSliceHookFunc(","),
		),
		Result: output,
	}
	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return dec.Decode(input)
}

// stringToWeakSliceHookFunc 与 viper 相同：字符串 → 任意 slice（逗号分隔）。
func stringToWeakSliceHookFunc(sep string) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Slice {
			return data, nil
		}
		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}
		return strings.Split(raw, sep), nil
	}
}
