package nacos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metlive/gohera/utils"
	"github.com/spf13/cast"
)

func TestParseAddr(t *testing.T) {
	addr, err := parseAddr("http://cloud-test.tal.com/tcm-api", 0)
	if err != nil {
		t.Fatal(err)
	}
	if addr.IpAddr != "cloud-test.tal.com" || addr.Port != 80 {
		t.Fatalf("unexpected addr: %+v", addr)
	}
	if addr.ContextPath != "/tcm-api/nacos" {
		t.Fatalf("contextPath=%q", addr.ContextPath)
	}

	addr2, err := parseAddr("127.0.0.1:8848", 9848)
	if err != nil {
		t.Fatal(err)
	}
	if addr2.Port != 8848 || addr2.GrpcPort != 9848 {
		t.Fatalf("unexpected addr2: %+v", addr2)
	}
}

func TestNormalizeConfigType(t *testing.T) {
	cases := map[string]struct {
		want string
		err  bool
	}{
		"yml":        {want: "yaml"},
		"YAML":       {want: "yaml"},
		"toml":       {want: "toml"},
		"json":       {want: "json"},
		"":           {want: "yaml"},
		"properties": {err: true},
	}
	for in, c := range cases {
		got, err := normalizeConfigType(in)
		if c.err {
			if err == nil {
				t.Fatalf("normalizeConfigType(%q) expected error", in)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Fatalf("normalizeConfigType(%q)=%q,%v want %q", in, got, err, c.want)
		}
	}
}

func TestApplyDefaults_DataIDByEnv(t *testing.T) {
	s := &Source{Env: func() string { return "dev" }}
	cfg := &bootstrapConfig{
		DataID:      "my-app",
		DataIDByEnv: true,
	}
	s.applyDefaults(cfg)
	if cfg.DataID != "my-app-dev" {
		t.Fatalf("dataId=%q", cfg.DataID)
	}
	if cfg.Mode != "http" {
		t.Fatalf("mode=%q", cfg.Mode)
	}
	// localPath 无默认值：本地值由引导文件非 nacos 段提供，仅显式配置时生效
	if cfg.LocalPath != "" {
		t.Fatalf("localPath=%q, want empty by default", cfg.LocalPath)
	}
}

func TestParseConfigContent_YAML(t *testing.T) {
	m, err := parseConfigContent("mysql:\n  dispatcher:\n    host: 127.0.0.1\n", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := utils.Lookup(m, "mysql.dispatcher.host")
	if !ok || cast.ToString(v) != "127.0.0.1" {
		t.Fatalf("host=%v ok=%v", v, ok)
	}
}

func TestParseConfigContent_Lowercase(t *testing.T) {
	m, err := parseConfigContent("HTTP:\n  Port: 8080\n", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := utils.Lookup(m, "http.port")
	if !ok || cast.ToInt(v) != 8080 {
		t.Fatalf("lowercase lookup failed: v=%v ok=%v", v, ok)
	}
}

func TestMergeEmptyContent(t *testing.T) {
	s := &Source{Merge: func(map[string]any) error {
		t.Fatal("merge should not be called for empty content")
		return nil
	}}
	if err := s.merge("   \n", "yaml"); err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestMergeInvalidYAML(t *testing.T) {
	s := &Source{Merge: func(map[string]any) error {
		t.Fatal("merge should not be called for invalid yaml")
		return nil
	}}
	if err := s.merge("::::not yaml\n[unterminated", "yaml"); err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "  ", "a", "b"); got != "a" {
		t.Fatalf("got %q", got)
	}
}

func writeBootstrapFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadBootstrapFile_EnvOverlay(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap.yaml",
		"nacos:\n  enabled: true\n  serveraddr: base.example\n  group: BASE_GROUP\n")
	writeBootstrapFile(t, dir, "bootstrap-test.yaml",
		"nacos:\n  serveraddr: test.example\n")

	s := &Source{Env: func() string { return "test" }}
	m, err := s.loadBootstrapFile([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := utils.Lookup(m, "nacos.serveraddr"); cast.ToString(v) != "test.example" {
		t.Fatalf("serveraddr=%v, want env overlay wins", v)
	}
	if v, _ := utils.Lookup(m, "nacos.enabled"); cast.ToBool(v) != true {
		t.Fatalf("enabled=%v, want base value kept", v)
	}
	if v, _ := utils.Lookup(m, "nacos.group"); cast.ToString(v) != "BASE_GROUP" {
		t.Fatalf("group=%v, want base value kept", v)
	}
}

func TestLoadBootstrapFile_EnvOnly(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap-prod.yaml",
		"nacos:\n  enabled: false\n  serveraddr: prod.example\n")

	s := &Source{Env: func() string { return "prod" }}
	m, err := s.loadBootstrapFile([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := utils.Lookup(m, "nacos.serveraddr"); cast.ToString(v) != "prod.example" {
		t.Fatalf("serveraddr=%v, want env-only file loaded", v)
	}
}

func TestLoadBootstrapFile_BaseOnly(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap.yaml",
		"nacos:\n  enabled: true\n  serveraddr: base.example\n")

	s := &Source{Env: func() string { return "prod" }} // 无 bootstrap-prod.yaml，回落基础文件
	m, err := s.loadBootstrapFile([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := utils.Lookup(m, "nacos.serveraddr"); cast.ToString(v) != "base.example" {
		t.Fatalf("serveraddr=%v, want base file as-is", v)
	}
}

func TestLoadBootstrapFile_MissThenHitNextDir(t *testing.T) {
	emptyDir := t.TempDir()
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap-dev.yaml",
		"nacos:\n  enabled: false\n")

	s := &Source{Env: func() string { return "dev" }}
	m, err := s.loadBootstrapFile([]string{emptyDir, dir})
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("want config from second search dir")
	}
	if v, _ := utils.Lookup(m, "nacos.enabled"); cast.ToBool(v) {
		t.Fatalf("enabled=%v, want false", v)
	}
}

func TestLoadBootstrapFile_None(t *testing.T) {
	s := &Source{Env: func() string { return "dev" }}
	m, err := s.loadBootstrapFile([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if m != nil {
		t.Fatalf("want nil map, got %v", m)
	}
}

func TestLoadBootstrapReturnsBody(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap-dev.yaml",
		"nacos:\n  enabled: false\n  serveraddr: dev.example\nmysql:\n  dispatcher:\n    host: 127.0.0.1\n")

	s := &Source{Env: func() string { return "dev" }, SearchPaths: []string{dir}}
	cfg, body, err := s.loadBootstrap()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("enabled should be false")
	}
	if cfg.ServerAddr != "dev.example" {
		t.Fatalf("serveraddr=%q", cfg.ServerAddr)
	}
	if v, ok := utils.Lookup(body, "mysql.dispatcher.host"); !ok || cast.ToString(v) != "127.0.0.1" {
		t.Fatalf("body should carry non-nacos sections: %v", body)
	}
	if _, ok := body["nacos"]; ok {
		t.Fatal("body must not include nacos section")
	}
}

func TestInitMergesBootstrapBody(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap-dev.yaml",
		"nacos:\n  enabled: false\nmysql:\n  dispatcher:\n    host: 127.0.0.1\n")

	var base, overlay map[string]any
	s := &Source{
		Env:         func() string { return "dev" },
		SearchPaths: []string{dir},
		Merge: func(m map[string]any) error {
			overlay = m
			return nil
		},
		MergeBase: func(m map[string]any) error {
			base = m
			return nil
		},
	}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if v, ok := utils.Lookup(base, "mysql.dispatcher.host"); !ok || cast.ToString(v) != "127.0.0.1" {
		t.Fatalf("MergeBase should receive mysql section: %v", base)
	}
	if overlay != nil {
		t.Fatalf("no local fallback file, overlay should stay nil: %v", overlay)
	}
}

func TestInitBodyIgnoredWithoutMergeBase(t *testing.T) {
	dir := t.TempDir()
	writeBootstrapFile(t, dir, "bootstrap-dev.yaml",
		"nacos:\n  enabled: false\nmysql:\n  a: 1\n")
	s := &Source{Env: func() string { return "dev" }, SearchPaths: []string{dir}}
	if err := s.Init(); err != nil {
		t.Fatal(err) // 未设置 MergeBase 时静默忽略 body（第三方直连用法向后兼容）
	}
}

func TestBootstrapExists(t *testing.T) {
	dir := t.TempDir()
	s := &Source{Env: func() string { return "prod" }, SearchPaths: []string{dir}}
	if s.BootstrapExists() {
		t.Fatal("no bootstrap file yet")
	}

	base := filepath.Join(dir, "bootstrap.yaml")
	if err := os.WriteFile(base, []byte("nacos: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !s.BootstrapExists() {
		t.Fatal("base bootstrap should count")
	}

	if err := os.Remove(base); err != nil {
		t.Fatal(err)
	}
	env := &Source{Env: func() string { return "prod" }, SearchPaths: []string{dir}}
	if err := os.WriteFile(filepath.Join(dir, "bootstrap-prod.yaml"), []byte("nacos: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !env.BootstrapExists() {
		t.Fatal("env bootstrap should count")
	}
	other := &Source{Env: func() string { return "test" }, SearchPaths: []string{dir}}
	if other.BootstrapExists() {
		t.Fatal("bootstrap-prod.yaml should not count for env=test")
	}
}
