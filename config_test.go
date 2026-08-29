package gohera

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/metlive/gohera/internal/configutil"
	"github.com/metlive/gohera/mysql"
	"github.com/metlive/gohera/redis"
)

func resetStore() {
	store.mu.Lock()
	if store.watchCancel != nil {
		store.watchCancel()
		store.watchCancel = nil
	}
	store.basePath = ""
	store.base = nil
	store.overlay = nil
	store.snapshot.Store(nil)
	store.mu.Unlock()
	store.subsMu.Lock()
	store.subs = nil
	store.subsMu.Unlock()
}

func setBase(m map[string]any) {
	store.mu.Lock()
	store.base = m
	store.overlay = nil
	store.snapshot.Store(store.buildLocked())
	store.mu.Unlock()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestGetTypes(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"http": map[string]any{
			"port":     8080,
			"enabled":  true,
			"timeout":  "5s",
			"rate":     1.5,
			"services": []any{"a", "b"},
		},
		"name": "demo",
	})
	if got := GetString("name"); got != "demo" {
		t.Fatalf("name=%q", got)
	}
	if got := GetInt("http.port"); got != 8080 {
		t.Fatalf("port=%d", got)
	}
	if !GetBool("http.enabled") {
		t.Fatal("enabled should be true")
	}
	if got := GetDuration("http.timeout"); got != 5*time.Second {
		t.Fatalf("timeout=%v", got)
	}
	if got := GetFloat64("http.rate"); got != 1.5 {
		t.Fatalf("rate=%v", got)
	}
	sl := GetStringSlice("http.services")
	if len(sl) != 2 || sl[0] != "a" || sl[1] != "b" {
		t.Fatalf("services=%v", sl)
	}
	if !IsSet("http.port") || IsSet("http.missing") {
		t.Fatal("IsSet mismatch")
	}
}

func TestUnmarshalKey_MySQL(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"mysql": map[string]any{
			"dispatcher": map[string]any{
				"host":           "127.0.0.1",
				"port":           3306,
				"user":           "root",
				"password":       "secret",
				"database":       "testdb",
				"max_life_time":  "30m",
				"max_open_conns": 100,
				"max_idle_conns": 20,
			},
		},
	})
	var cfg mysql.Config
	if err := UnmarshalKey("mysql.dispatcher", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 3306 || cfg.User != "root" || cfg.Database != "testdb" {
		t.Fatalf("cfg=%+v", cfg)
	}
	if cfg.MaxLifeTime != 30*time.Minute || cfg.MaxOpenConns != 100 || cfg.MaxIdleConns != 20 {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestUnmarshalKey_Redis(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"redis": map[string]any{
			"address":      "127.0.0.1:6379",
			"auth":         "mypass",
			"database":     1,
			"max_idle":     10,
			"idle_timeout": "30s",
		},
	})
	var cfg redis.Config
	if err := UnmarshalKey("redis", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "127.0.0.1:6379" || cfg.Auth != "mypass" || cfg.Database != 1 {
		t.Fatalf("cfg=%+v", cfg)
	}
	if cfg.MaxIdle != 10 || cfg.IdleTimeout != 30*time.Second {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	resetStore()
	t.Setenv("APP_HTTP_PORT", "9999")
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	if got := GetString("http.port"); got != "9999" {
		t.Fatalf("port=%q", got)
	}
}

func TestEnvOverridesOverlay(t *testing.T) {
	resetStore()
	t.Setenv("APP_HTTP_PORT", "7777")
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	store.applyOverlay(map[string]any{"http": map[string]any{"port": 9090}})
	if got := GetString("http.port"); got != "7777" {
		t.Fatalf("env should beat overlay: %q", got)
	}
}

func TestEnvUnmarshalKeySubtree(t *testing.T) {
	resetStore()
	t.Setenv("APP_MYSQL_DEFAULT_PASSWORD", "envpass")
	setBase(map[string]any{
		"mysql": map[string]any{
			"default": map[string]any{
				"host": "1.1.1.1", "user": "u", "database": "d", "password": "filepass",
			},
		},
	})
	var cfg mysql.Config
	if err := UnmarshalKey("mysql.default", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Password != "envpass" {
		t.Fatalf("password=%q want envpass", cfg.Password)
	}
}

func TestEnvDoesNotOverwriteMapNode(t *testing.T) {
	resetStore()
	t.Setenv("APP_MYSQL_DISPATCHER", "should-drop")
	setBase(map[string]any{
		"mysql": map[string]any{
			"dispatcher": map[string]any{"host": "1.1.1.1", "user": "u", "database": "d"},
		},
	})
	if GetString("mysql.dispatcher.host") != "1.1.1.1" {
		t.Fatalf("map node overwritten: host=%q", GetString("mysql.dispatcher.host"))
	}
	if GetString("mysql.dispatcher") == "should-drop" {
		t.Fatal("env overwrote map node with string")
	}
}

func TestOverlayBranchPreservation(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"mysql": map[string]any{
			"other":      map[string]any{"host": "10.0.0.1"},
			"dispatcher": map[string]any{"host": "10.0.0.2"},
		},
	})
	store.applyOverlay(map[string]any{
		"mysql": map[string]any{
			"dispatcher": map[string]any{"host": "10.0.0.3"},
		},
	})
	if got := GetString("mysql.other.host"); got != "10.0.0.1" {
		t.Fatalf("other.host=%q (branch lost)", got)
	}
	if got := GetString("mysql.dispatcher.host"); got != "10.0.0.3" {
		t.Fatalf("dispatcher.host=%q", got)
	}
}

func TestApplyOverlayIsolatedFromCaller(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	src := map[string]any{"http": map[string]any{"port": 9090}}
	store.applyOverlay(src)
	src["http"].(map[string]any)["port"] = 1
	if GetInt("http.port") != 9090 {
		t.Fatalf("overlay shared with caller: %d", GetInt("http.port"))
	}
}

func TestMultiDBEnvNoPseudoEntry(t *testing.T) {
	resetStore()
	t.Setenv("APP_MYSQL_DEFAULT_PASSWORD", "x")
	setBase(map[string]any{
		"mysql": map[string]any{
			"main": map[string]any{"host": "1.1.1.1", "user": "u", "database": "d", "password": "p"},
		},
	})
	m := GetStringMap("mysql")
	if _, ok := m["default"]; ok {
		t.Fatalf("pseudo DB entry created: %v", m)
	}
	if len(m) != 1 {
		t.Fatalf("mysql map=%v", m)
	}
}

func TestEnvUnderscoreDBNotPseudo(t *testing.T) {
	resetStore()
	t.Setenv("APP_MYSQL_USER_DB_PASSWORD", "x")
	setBase(map[string]any{
		"mysql": map[string]any{
			"user_db": map[string]any{"host": "1.1.1.1", "user": "u", "database": "d", "password": "p"},
		},
	})
	m := GetStringMap("mysql")
	if _, ok := m["user"]; ok {
		t.Fatalf("pseudo hierarchy created: %v", m)
	}
	if _, ok := m["user_db"]; !ok {
		t.Fatalf("user_db missing: %v", m)
	}
}

func TestGetterReadOnly(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"mysql": map[string]any{
			"main": map[string]any{"host": "1.1.1.1", "password": "p"},
		},
		"list":     []any{"a", "b"},
		"str_list": []string{"a", "b"},
	})

	m := GetStringMap("mysql")
	m["hacked"] = "yes"
	if inner, ok := m["main"].(map[string]any); ok {
		inner["host"] = "hacked"
	}
	if GetString("mysql.main.host") != "1.1.1.1" {
		t.Fatalf("snapshot polluted: host=%q", GetString("mysql.main.host"))
	}
	if IsSet("mysql.hacked") {
		t.Fatal("snapshot polluted with hacked key")
	}

	if raw := GetConfig("mysql"); raw != nil {
		if rm, ok := raw.(map[string]any); ok {
			rm["hacked2"] = true
		}
	}
	if IsSet("mysql.hacked2") {
		t.Fatal("GetConfig returned shared reference")
	}

	sl := GetStringSlice("list")
	sl[0] = "z"
	if got := GetStringSlice("list"); got[0] != "a" {
		t.Fatalf("slice polluted: %v", got)
	}

	sl2 := GetStringSlice("str_list")
	sl2[0] = "z"
	if got := GetStringSlice("str_list"); got[0] != "a" {
		t.Fatalf("[]string slice polluted: %v", got)
	}
}

func TestCaseInsensitive(t *testing.T) {
	resetStore()
	setBase(map[string]any{"HTTP": map[string]any{"Port": 8080}})
	if got := GetInt("http.port"); got != 8080 {
		t.Fatalf("uppercase key lookup failed: %d", got)
	}
	if got := GetInt("HTTP.PORT"); got != 8080 {
		t.Fatalf("uppercase query lookup failed: %d", got)
	}
}

func TestOverlayCaseInsensitive(t *testing.T) {
	resetStore()
	setBase(map[string]any{"db": map[string]any{"host": "base"}})
	store.applyOverlay(map[string]any{"DB": map[string]any{"HOST": "overlay"}})
	if got := GetString("db.host"); got != "overlay" {
		t.Fatalf("overlay override failed: %q", got)
	}
}

func TestAppMetaEnvIsolation(t *testing.T) {
	resetStore()
	t.Setenv("APP_CONFIG", "x")
	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_MODE", "prod")
	t.Setenv("APP_NAME", "app")
	t.Setenv("APP_VERSION", "1.0")
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	for _, k := range []string{"config", "env", "mode", "name", "version"} {
		if IsSet(k) {
			t.Fatalf("meta var leaked into snapshot: %s", k)
		}
	}
}

func TestOnConfigChange(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	called := make(chan struct{}, 1)
	OnConfigChange(func() { called <- struct{}{} })
	store.applyOverlay(map[string]any{"http": map[string]any{"port": 9090}})
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("subscriber not called")
	}
	if GetInt("http.port") != 9090 {
		t.Fatalf("port=%d", GetInt("http.port"))
	}
}

func TestOnConfigChangeNotOnFailure(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	store.mu.Lock()
	store.basePath = "/nonexistent/app.yaml"
	store.mu.Unlock()
	called := make(chan struct{}, 1)
	OnConfigChange(func() { called <- struct{}{} })
	store.reloadFromFile()
	select {
	case <-called:
		t.Fatal("subscriber should not be called on failure")
	case <-time.After(500 * time.Millisecond):
	}
	if GetInt("http.port") != 8080 {
		t.Fatalf("snapshot should be preserved: %d", GetInt("http.port"))
	}
}

func TestOnConfigChangePanicSafe(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	second := make(chan struct{}, 1)
	OnConfigChange(func() { panic("boom") })
	OnConfigChange(func() { second <- struct{}{} })
	store.applyOverlay(map[string]any{"http": map[string]any{"port": 9090}})
	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatal("second subscriber not called after panic")
	}
}

func TestUnmarshalWhole(t *testing.T) {
	resetStore()
	setBase(map[string]any{
		"http": map[string]any{"port": 8080, "service": "demo"},
	})
	var cfg struct {
		HTTP struct {
			Port    int    `mapstructure:"port"`
			Service string `mapstructure:"service"`
		} `mapstructure:"http"`
	}
	if err := Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Port != 8080 || cfg.HTTP.Service != "demo" {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestUnmarshalKeyMissing(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	var cfg mysql.Config
	if err := UnmarshalKey("mysql.nope", &cfg); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestUnmarshalInvalidPtr(t *testing.T) {
	if err := Unmarshal(nil); err == nil {
		t.Fatal("expected error for nil")
	}
	var v int
	if err := UnmarshalKey("x", v); err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestFileReloadKeepsOverlay(t *testing.T) {
	resetStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	writeFile(t, path, "http:\n  port: 8080\n")

	base, err := configutil.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.basePath = path
	store.base = base
	store.snapshot.Store(store.buildLocked())
	store.mu.Unlock()

	store.applyOverlay(map[string]any{"http": map[string]any{"timeout": "5s"}})
	if !IsSet("http.timeout") {
		t.Fatal("overlay not applied")
	}

	writeFile(t, path, "http:\n  port: 9090\n")
	store.reloadFromFile()
	if GetInt("http.port") != 9090 {
		t.Fatalf("port=%d", GetInt("http.port"))
	}
	if !IsSet("http.timeout") {
		t.Fatal("overlay lost after file reload")
	}
}

func TestReloadInvalidYAMLKeepsSnapshot(t *testing.T) {
	resetStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	writeFile(t, path, "http:\n  port: 8080\n")

	base, err := configutil.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.basePath = path
	store.base = base
	store.snapshot.Store(store.buildLocked())
	store.mu.Unlock()

	called := make(chan struct{}, 1)
	OnConfigChange(func() { called <- struct{}{} })
	writeFile(t, path, "::::not yaml\n[unterminated")
	store.reloadFromFile()
	select {
	case <-called:
		t.Fatal("subscriber should not fire on parse failure")
	case <-time.After(300 * time.Millisecond):
	}
	if GetInt("http.port") != 8080 {
		t.Fatalf("snapshot lost: %d", GetInt("http.port"))
	}
}

func TestWatchRenameAndRemoveCreate(t *testing.T) {
	resetStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	writeFile(t, path, "http:\n  port: 8080\n")

	base, err := configutil.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.basePath = path
	store.base = base
	store.snapshot.Store(store.buildLocked())
	store.mu.Unlock()
	store.startWatch(path)
	t.Cleanup(resetStore)

	tmp := filepath.Join(t.TempDir(), "tmp.yaml")
	writeFile(t, tmp, "http:\n  port: 9090\n")
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return GetInt("http.port") == 9090 })

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, "http:\n  port: 7070\n")
	waitFor(t, func() bool { return GetInt("http.port") == 7070 })
}

func TestWatchBadFileKeepsSnapshot(t *testing.T) {
	resetStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	writeFile(t, path, "http:\n  port: 8080\n")

	base, err := configutil.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.basePath = path
	store.base = base
	store.snapshot.Store(store.buildLocked())
	store.mu.Unlock()
	store.startWatch(path)
	t.Cleanup(resetStore)

	called := make(chan struct{}, 1)
	OnConfigChange(func() { called <- struct{}{} })
	writeFile(t, path, "::::bad")
	select {
	case <-called:
		t.Fatal("subscriber should not fire on bad file")
	case <-time.After(800 * time.Millisecond):
	}
	if GetInt("http.port") != 8080 {
		t.Fatalf("snapshot lost after bad watch write: %d", GetInt("http.port"))
	}
}

func TestDiscoverAPPConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	writeFile(t, path, "http:\n  port: 1\n")
	t.Setenv("APP_CONFIG_FILE", path)
	t.Setenv("APP_CONFIG", "")
	got, err := discoverConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("got %q want %q", got, path)
	}
}

func TestDiscoverPreferredName(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, "myapp.yaml"), "http:\n  port: 2\n")
	t.Setenv("APP_CONFIG_FILE", "")
	t.Setenv("APP_CONFIG", "myapp")
	got, err := discoverConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "myapp.yaml" {
		t.Fatalf("got %q", got)
	}
}

func TestDiscoverMultipleFilesError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, "a.yaml"), "a: 1\n")
	writeFile(t, filepath.Join(dir, "b.yaml"), "b: 1\n")
	t.Setenv("APP_CONFIG_FILE", "")
	t.Setenv("APP_CONFIG", "")
	_, err := discoverConfigFile()
	if err == nil {
		t.Fatal("expected multiple file error")
	}
}

func TestSoakConcurrentReadRebuild(t *testing.T) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"host": "127.0.0.1", "port": 8080}})

	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = GetString("http.host")
					_ = GetInt("http.port")
					_ = IsSet("http.port")
					_ = GetStringMap("http")
					_ = GetStringSlice("http.host")
				}
			}
		}()
	}
	for i := 0; i < 80; i++ {
		store.applyOverlay(map[string]any{"http": map[string]any{"port": 9000 + i%10}})
		setBase(map[string]any{"http": map[string]any{"host": "127.0.0.1", "port": 8080}})
	}
	close(done)
	wg.Wait()
	if !IsSet("http.host") {
		t.Fatal("host missing after soak")
	}
}

func BenchmarkGetString(b *testing.B) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"host": "127.0.0.1"}})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetString("http.host")
	}
}

func BenchmarkGetInt(b *testing.B) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetInt("http.port")
	}
}

func BenchmarkIsSet(b *testing.B) {
	resetStore()
	setBase(map[string]any{"http": map[string]any{"port": 8080}})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsSet("http.port")
	}
}
