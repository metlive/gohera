package nacos

import (
	"strings"
	"testing"
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
	cases := map[string]string{
		"yml":  "yaml",
		"YAML": "yaml",
		"toml": "toml",
		"":     "yaml",
	}
	for in, want := range cases {
		if got := normalizeConfigType(in); got != want {
			t.Fatalf("normalizeConfigType(%q)=%q want %q", in, got, want)
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
	v, err := parseConfigContent("mysql:\n  dispatcher:\n    host: 127.0.0.1\n", "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if v.GetString("mysql.dispatcher.host") != "127.0.0.1" {
		t.Fatalf("host=%q", v.GetString("mysql.dispatcher.host"))
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "  ", "a", "b"); got != "a" {
		t.Fatalf("got %q", got)
	}
}
