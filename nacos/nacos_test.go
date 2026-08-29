package nacos

import (
	"strings"
	"testing"

	"github.com/metlive/gohera/internal/configutil"
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
	if !strings.Contains(cfg.LocalPath, "nacos.dev.yaml") {
		t.Fatalf("localPath=%q", cfg.LocalPath)
	}
}

func TestParseConfigContent_YAML(t *testing.T) {
	m, err := parseConfigContent("mysql:\n  dispatcher:\n    host: 127.0.0.1\n", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := configutil.Lookup(m, "mysql.dispatcher.host")
	if !ok || cast.ToString(v) != "127.0.0.1" {
		t.Fatalf("host=%v ok=%v", v, ok)
	}
}

func TestParseConfigContent_Lowercase(t *testing.T) {
	m, err := parseConfigContent("HTTP:\n  Port: 8080\n", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := configutil.Lookup(m, "http.port")
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
