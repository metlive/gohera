package gohera

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/maps"
	"github.com/metlive/gohera/internal/configutil"
)

// configSnapshot 是一次 rebuild 发布的只读配置视图：
// flat 供 Get* / IsSet 精确查找，nested 供 Unmarshal / UnmarshalKey 解码。
type configSnapshot struct {
	flat   map[string]any
	nested map[string]any
}

// configStore 持有三层输入（本地文件 < overlay < env），每次变更整表重建后原子发布。
type configStore struct {
	mu       sync.RWMutex
	basePath string
	base     map[string]any
	overlay  map[string]any

	snapshot atomic.Pointer[configSnapshot]

	subsMu sync.Mutex
	subs   []func()

	watchCancel context.CancelFunc
}

var store = &configStore{}

// init 加载本地配置文件并启动目录级 Watch（仅首次成功加载后）。
func (s *configStore) init(path string) error {
	base, err := configutil.LoadFile(path)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.basePath = path
	s.base = base
	s.snapshot.Store(s.buildLocked())
	s.mu.Unlock()

	s.startWatch(path)
	return nil
}

// initEmpty 以空 base 初始化（无 app 配置文件、仅依赖 bootstrap/Nacos 提供配置时使用），
// 不启动文件 watch；后续 overlay 合并照常生效。
func (s *configStore) initEmpty() error {
	s.mu.Lock()
	s.basePath = ""
	s.base = map[string]any{}
	s.snapshot.Store(s.buildLocked())
	s.mu.Unlock()
	return nil
}

// mergeBase 将环境级本地配置（bootstrap-{env}.yaml 非 nacos 段）深合并进 base 层：
// 覆盖 app.yaml 同名键、低于 overlay（远程 Nacos / 兜底文件），重建快照并通知订阅者。
func (s *configStore) mergeBase(m map[string]any) {
	if len(m) == 0 {
		return
	}
	s.mu.Lock()
	if s.base == nil {
		s.base = map[string]any{}
	}
	maps.Merge(configutil.Lowercase(m), s.base)
	s.snapshot.Store(s.buildLocked())
	s.mu.Unlock()
	s.notify()
}

// applyOverlay 用 Nacos/兜底配置替换 overlay 层并重建快照。
func (s *configStore) applyOverlay(overlay map[string]any) {
	s.mu.Lock()
	if overlay == nil {
		s.overlay = nil
	} else {
		s.overlay = configutil.Lowercase(overlay)
	}
	s.snapshot.Store(s.buildLocked())
	s.mu.Unlock()
	s.notify()
}

// reloadFromFile 文件变更后重读本地配置并重建；失败保留上一份快照且不通知。
func (s *configStore) reloadFromFile() {
	s.mu.RLock()
	path := s.basePath
	s.mu.RUnlock()
	if path == "" {
		return
	}

	base, err := configutil.LoadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[gohera] config reload failed: %v\n", err)
		return
	}

	s.mu.Lock()
	s.base = base
	s.snapshot.Store(s.buildLocked())
	s.mu.Unlock()
	s.notify()
}

// buildLocked 三层合并（base < overlay < env）并生成快照。调用方需持有写锁。
// configutil.Lowercase 同时完成深拷贝与 key 归一化，保证快照与源层隔离。
func (s *configStore) buildLocked() *configSnapshot {
	nested := configutil.Lowercase(s.base)
	maps.Merge(configutil.Lowercase(s.overlay), nested)
	applyEnv(nested)
	return &configSnapshot{flat: flattenMap(nested), nested: nested}
}

// lookup 热路径：读原子快照，key 一次 lowercase 后精确查 flat。
func (s *configStore) lookup(key string) (any, bool) {
	snap := s.snapshot.Load()
	if snap == nil {
		return nil, false
	}
	v, ok := snap.flat[strings.ToLower(key)]
	return v, ok
}

// notify 异步、panic-safe 地按注册顺序调用订阅者，不阻塞 rebuild 调用方。
func (s *configStore) notify() {
	s.subsMu.Lock()
	subs := make([]func(), len(s.subs))
	copy(subs, s.subs)
	s.subsMu.Unlock()
	if len(subs) == 0 {
		return
	}

	go func() {
		for _, fn := range subs {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Fprintf(os.Stderr, "[gohera] config subscriber panic: %v\n", r)
					}
				}()
				fn()
			}()
		}
	}()
}

// startWatch 目录级 watch：Remove 不退出、创建失败不 os.Exit。
func (s *configStore) startWatch(path string) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.watchCancel != nil {
		s.watchCancel()
	}
	s.watchCancel = cancel
	s.mu.Unlock()

	go s.runWatch(ctx, path)
}

func (s *configStore) runWatch(ctx context.Context, path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[gohera] config watch disabled: %v\n", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(filepath.Dir(path)); err != nil {
		fmt.Fprintf(os.Stderr, "[gohera] config watch disabled: %v\n", err)
		return
	}

	s.watchLoop(ctx, watcher, path)
}

func (s *configStore) watchLoop(ctx context.Context, watcher *fsnotify.Watcher, path string) {
	name := filepath.Base(path)
	realPath, _ := filepath.EvalSymlinks(path)
	var timer *time.Timer
	schedule := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(100*time.Millisecond, s.reloadFromFile)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(ev.Name) == name {
				schedule()
				continue
			}
			// k8s ConfigMap：..data symlink 切换后实路径变化
			if cur, err := filepath.EvalSymlinks(path); err == nil && cur != "" && cur != realPath {
				realPath = cur
				schedule()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "[gohera] config watch error: %v\n", err)
		}
	}
}

// appMetaEnvVars 与应用元信息相关的 APP_*（去前缀并 lowercase 后），
// 不参与配置覆盖，避免污染 config/mode/name/version 段。
var appMetaEnvVars = map[string]struct{}{
	"config":      {}, // APP_CONFIG
	"config_file": {}, // APP_CONFIG_FILE
	"env":         {}, // APP_ENV（部署级别，由 parseEnv 直读）
	"mode":        {}, // APP_MODE
	"name":        {}, // APP_NAME
	"version":     {}, // APP_VERSION
}

// applyEnv 按 reconcile 规则应用 APP_*：仅覆盖 base ∪ overlay 中已存在的叶子。
func applyEnv(nested map[string]any) {
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(k, "APP_") {
			continue
		}
		key := strings.ToLower(strings.TrimPrefix(k, "APP_"))
		if _, meta := appMetaEnvVars[key]; meta {
			continue
		}
		setEnv(nested, strings.ReplaceAll(key, "_", "."), v)
	}
}

// setEnv 仅当 key 路径已存在时写叶子，否则丢弃（不新增子树层级/伪条目）。
func setEnv(nested map[string]any, key, val string) {
	parts := strings.Split(key, ".")
	cur := nested
	for i, p := range parts {
		next, ok := cur[p]
		if !ok {
			return
		}
		if i == len(parts)-1 {
			if _, isMap := next.(map[string]any); isMap {
				return
			}
			cur[p] = val
			return
		}
		m, ok := next.(map[string]any)
		if !ok {
			return
		}
		cur = m
	}
}

// flattenMap 递归扁平化，同时保留中间 map 节点与叶子（对齐现网 flattenSettings）。
func flattenMap(m map[string]any) map[string]any {
	flat := make(map[string]any)
	var walk func(prefix string, mm map[string]any)
	walk = func(prefix string, mm map[string]any) {
		for k, v := range mm {
			full := k
			if prefix != "" {
				full = prefix + "." + k
			}
			flat[full] = v
			if sub, ok := v.(map[string]any); ok {
				walk(full, sub)
			}
		}
	}
	walk("", m)
	return flat
}

func deepCopyValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return deepCopyMap(t)
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = deepCopyValue(t[i])
		}
		return out
	case []string:
		out := make([]string, len(t))
		copy(out, t)
		return out
	default:
		return v
	}
}

func deepCopyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = deepCopyValue(v)
	}
	return out
}

// OnConfigChange 注册配置变更回调：每次 rebuild 成功后异步调用，订阅不回溯。
func OnConfigChange(fn func()) {
	if fn == nil {
		return
	}
	store.subsMu.Lock()
	store.subs = append(store.subs, fn)
	store.subsMu.Unlock()
}
