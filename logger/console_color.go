package logger

import (
	"os"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"
)

const (
	ansiReset   = "\033[0m"
	ansiCyan    = "\033[36m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiRed     = "\033[31m"
	ansiBoldRed = "\033[1;31m"
)

// consoleColorEnabled 检测 stdout 是否为终端且未设置 NO_COLOR。
func consoleColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}

// colorizeLevel 为控制台日志级别添加 ANSI 颜色，非 TTY 时原样返回。
func colorizeLevel(level zapcore.Level, useColor bool) string {
	s := level.CapitalString()
	if !useColor {
		return s
	}
	switch level {
	case zapcore.DebugLevel:
		return ansiCyan + s + ansiReset
	case zapcore.InfoLevel:
		return ansiGreen + s + ansiReset
	case zapcore.WarnLevel:
		return ansiYellow + s + ansiReset
	case zapcore.ErrorLevel:
		return ansiRed + s + ansiReset
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ansiBoldRed + s + ansiReset
	default:
		return s
	}
}
