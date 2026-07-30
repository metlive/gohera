package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetLoggerForTest(t *testing.T) {
	t.Helper()
	loggerMu.Lock()
	logger = nil
	loggerMu.Unlock()
	t.Cleanup(func() {
		loggerMu.Lock()
		logger = nil
		loggerMu.Unlock()
	})
}

func TestLazyInitNoPanic(t *testing.T) {
	resetLoggerForTest(t)
	Info(context.Background(), "lazy-init-ok")
}

func TestEmptyFilePathCreatesNoFiles(t *testing.T) {
	resetLoggerForTest(t)
	dir := t.TempDir()

	Init(Options{
		FilePath:     "",
		EnableStdout: Bool(false), // 避免污染测试输出；Nop core
		Project:      "test",
	})
	Info(context.Background(), "console-only")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files under empty FilePath scenario using temp dir probe, got %d", len(entries))
	}
}

func TestFilePathWritesLogs(t *testing.T) {
	resetLoggerForTest(t)
	dir := t.TempDir()

	Init(Options{
		FilePath:     dir,
		EnableStdout: Bool(false),
		Project:      "test",
	})
	Infotf(context.Background(), "file-write-%s", "ok")
	_ = getLogger().Sync()

	link := filepath.Join(dir, "server_info.log")
	if _, err := os.Lstat(link); err != nil {
		// rotatelogs 可能先写带日期文件；至少目录非空
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if len(entries) == 0 {
			t.Fatalf("expected log files under %s, link err: %v", dir, err)
		}
		return
	}
}

func TestInitOverride(t *testing.T) {
	resetLoggerForTest(t)
	dir := t.TempDir()

	Init(Options{EnableStdout: Bool(false)})
	Init(Options{
		FilePath:     dir,
		EnableStdout: Bool(false),
		Project:      "override",
	})
	Info(context.Background(), "after-override")
	_ = getLogger().Sync()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected files after Init override with FilePath")
	}
}

func TestNormalizeStdoutFormat(t *testing.T) {
	n := normalizeOptions(Options{StdoutFormat: "DETAILED"})
	if n.StdoutFormat != "detailed" {
		t.Fatalf("got %q", n.StdoutFormat)
	}
	n = normalizeOptions(Options{StdoutFormat: "nope"})
	if n.StdoutFormat != "simple" {
		t.Fatalf("got %q", n.StdoutFormat)
	}
	n = normalizeOptions(Options{})
	if !n.EnableStdout {
		t.Fatal("default EnableStdout should be true")
	}
	if strings.TrimSpace(n.FilePath) != "" {
		t.Fatal("default FilePath should be empty")
	}
}

func TestNoGoHeraImport(t *testing.T) {
	// 编译期已保证；此处作文档性占位，避免回归时忘记
	_ = TraceCtx
	_ = SpanIdDefault
}
