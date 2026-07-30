package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/metlive/gohera/logger/internal/rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Context / span 常量（与 gohera 历史字面量保持一致，便于独立使用）。
const (
	TraceCtx      = "trace-ctx"
	SpanIdDefault = "1"
)

type Trace struct {
	TraceId string `json:"trace_id"`
	SpanId  string `json:"span_id"`
	UserId  int    `json:"user_id"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Status  int    `json:"status"`
	Headers map[string]any
}

// Options 控制日志输出目标与格式。
type Options struct {
	// FilePath 日志目录。空字符串：不写任何日志文件。
	// 非空时用 filepath.Join(FilePath, "server_{level}.log")。
	FilePath string

	// EnableStdout 是否输出到控制台。
	// nil → 默认 true；非 nil → 按 *EnableStdout。
	EnableStdout *bool

	// StdoutFormat: "simple"（默认）| "detailed"
	StdoutFormat string

	// Project 写入全局字段 x_project；空则写空字符串。
	Project string
}

// Bool 便于填写 Options.EnableStdout。
func Bool(v bool) *bool { return &v }

var (
	logger   *zap.Logger
	loggerMu sync.Mutex
)

// Init 按 Options 配置（或重配）包级 logger。可重复调用；后一次覆盖前一次。
func Init(opts Options) {
	n := normalizeOptions(opts)
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger = buildLogger(n)
}

// InitLogger 兼容入口：logPath 非空则写文件；空路径则仅控制台。
func InitLogger(logPath string) {
	Init(Options{
		FilePath:     logPath,
		EnableStdout: Bool(true),
	})
}

// InitLoggerWithStdout 初始化日志，并控制是否输出到控制台。
func InitLoggerWithStdout(logPath string, enableStdout bool) {
	Init(Options{
		FilePath:     logPath,
		EnableStdout: Bool(enableStdout),
	})
}

// InitLoggerWithStdoutFormat 初始化日志，支持控制台输出及格式。
func InitLoggerWithStdoutFormat(logPath string, enableStdout bool, stdoutFormat string) {
	Init(Options{
		FilePath:     logPath,
		EnableStdout: Bool(enableStdout),
		StdoutFormat: stdoutFormat,
	})
}

type normalizedOptions struct {
	FilePath     string
	EnableStdout bool
	StdoutFormat string
	Project      string
}

func normalizeOptions(opts Options) normalizedOptions {
	out := normalizedOptions{
		FilePath:     strings.TrimSpace(opts.FilePath),
		EnableStdout: true,
		StdoutFormat: "simple",
		Project:      opts.Project,
	}
	if opts.EnableStdout != nil {
		out.EnableStdout = *opts.EnableStdout
	}
	switch strings.ToLower(strings.TrimSpace(opts.StdoutFormat)) {
	case "", "simple":
		out.StdoutFormat = "simple"
	case "detailed":
		out.StdoutFormat = "detailed"
	default:
		fmt.Fprintf(os.Stderr, "[logger] unknown StdoutFormat %q, fallback to simple\n", opts.StdoutFormat)
		out.StdoutFormat = "simple"
	}
	return out
}

func buildLogger(opts normalizedOptions) *zap.Logger {
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	cores := make([]zapcore.Core, 0, 5)
	if opts.FilePath != "" {
		cores = append(cores,
			getEncoderCore(filepath.Join(opts.FilePath, "server_debug.log"), debugPriority),
			getEncoderCore(filepath.Join(opts.FilePath, "server_info.log"), infoPriority),
			getEncoderCore(filepath.Join(opts.FilePath, "server_warn.log"), warnPriority),
			getEncoderCore(filepath.Join(opts.FilePath, "server_error.log"), errorPriority),
		)
	}

	if opts.EnableStdout {
		if opts.StdoutFormat == "detailed" {
			cores = append(cores, getDetailedConsoleCore(zap.DebugLevel))
		} else {
			cores = append(cores, getConsoleCore(zap.DebugLevel))
		}
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.Fields(
		zap.String("x_type", "go"),
		zap.String("x_project", opts.Project),
	)).WithOptions(zap.AddCallerSkip(1))
}

func getLogger() *zap.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if logger == nil {
		logger = buildLogger(normalizeOptions(Options{}))
	}
	return logger
}

func getConsoleCore(level zapcore.LevelEnabler) zapcore.Core {
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
	}), zapcore.AddSync(os.Stdout), level)
	return &cleanConsoleCore{Core: core}
}

func getDetailedConsoleCore(level zapcore.LevelEnabler) zapcore.Core {
	return &detailedConsoleCore{
		useColor: consoleColorEnabled(),
		Core: zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				TimeKey:        "T",
				LevelKey:       "L",
				NameKey:        "N",
				CallerKey:      "C",
				MessageKey:     "M",
				StacktraceKey:  "S",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     detailedTimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
				EncodeName:     zapcore.FullNameEncoder,
			}),
			zapcore.AddSync(os.Stdout),
			level,
		),
	}
}

func detailedTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + t.Format("2006-01-02 15:04:05.000") + "]")
}

type detailedConsoleCore struct {
	zapcore.Core
	useColor bool
}

func (c *detailedConsoleCore) With(fields []zapcore.Field) zapcore.Core {
	return &detailedConsoleCore{Core: c.Core.With(fields), useColor: c.useColor}
}

func (c *detailedConsoleCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *detailedConsoleCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf := make([]byte, 0, 256)

	buf = append(buf, '[')
	buf = ent.Time.AppendFormat(buf, "2006-01-02 15:04:05.000")
	buf = append(buf, ']')

	buf = append(buf, ' ')
	buf = append(buf, colorizeLevel(ent.Level, c.useColor)...)

	if ent.Caller.Defined {
		buf = append(buf, ' ', '[')
		buf = append(buf, ent.Caller.TrimmedPath()...)
		buf = append(buf, ']')
	}

	if traceID := extractTraceID(fields); traceID != "" {
		buf = append(buf, ' ', '[')
		buf = append(buf, traceID...)
		buf = append(buf, ']')
	}

	buf = append(buf, ' ', ':', ' ')
	buf = append(buf, ent.Message...)
	buf = append(buf, '\n')

	_, err := os.Stdout.Write(buf)
	return err
}

func extractTraceID(fields []zapcore.Field) string {
	for _, f := range fields {
		if f.Key == "x_trace_id" {
			return f.String
		}
	}
	return ""
}

type cleanConsoleCore struct {
	zapcore.Core
}

func (c *cleanConsoleCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *cleanConsoleCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *cleanConsoleCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(ent, nil)
}

func getEncoderCore(fileName string, level zapcore.LevelEnabler) zapcore.Core {
	logf, err := rotatelogs.New(fileName+"_%Y-%m-%d",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[logger] failed to create rotatelogs for %s: %v\n", fileName, err)
		panic(fmt.Errorf("failed to create log file %s: %w", fileName, err))
	}

	writer := zapcore.AddSync(logf)
	return zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:    "x_message",
		LevelKey:      "x_level",
		StacktraceKey: "x_trace",
		TimeKey:       "x_time",
		NameKey:       "x_logger",
		CallerKey:     "x_caller",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}), writer, level)
}

// GetTraceContext 从 Context 中获取 Trace 信息。
func GetTraceContext(ctx context.Context) *Trace {
	if ctx == nil {
		return new(Trace)
	}
	if v := ctx.Value(TraceCtx); v != nil {
		if t, ok := v.(*Trace); ok {
			return t
		}
	}
	return new(Trace)
}

func getContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		ctx = context.Background()
	}
	traceInfo := GetTraceContext(ctx)
	traceID := traceInfo.TraceId
	if traceID == "" {
		traceID = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	spanID := traceInfo.SpanId
	if spanID == "" {
		spanID = SpanIdDefault
	}
	return []zap.Field{
		zap.String("x_trace_id", traceID),
		zap.String("x_span_id", spanID),
		zap.Int("x_user_id", traceInfo.UserId),
		zap.String("x_path", traceInfo.Path),
		zap.Int("x_status", traceInfo.Status),
		zap.Any("x_header", maskSensitiveHeaders(traceInfo.Headers)),
	}
}

// StartSpan 处理日志格式化并从 Context 中提取跟踪信息。
func StartSpan(ctx context.Context, format string, args ...any) (string, []zap.Field) {
	l := len(args)
	if l > 0 {
		return fmt.Sprintf(format, args[:l]...), getContextFields(ctx)
	}
	return format, getContextFields(ctx)
}

func Info(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	getLogger().Info(str, filed...)
}

func Infotf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	getLogger().Info(str, filed...)
}

func Warn(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	getLogger().Warn(str, filed...)
}

func Warntf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	getLogger().Warn(str, filed...)
}

func Error(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	getLogger().Error(str, filed...)
}

func Errortf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	getLogger().Error(str, filed...)
}

type ContextLogger struct {
	ctx context.Context
}

func Ctx(ctx context.Context) *ContextLogger {
	return &ContextLogger{ctx: ctx}
}

func (l *ContextLogger) Info(args ...any) {
	Info(l.ctx, args...)
}

func (l *ContextLogger) Infotf(template string, args ...any) {
	Infotf(l.ctx, template, args...)
}

func (l *ContextLogger) Warn(args ...any) {
	Warn(l.ctx, args...)
}

func (l *ContextLogger) Warntf(template string, args ...any) {
	Warntf(l.ctx, template, args...)
}

func (l *ContextLogger) Error(args ...any) {
	Error(l.ctx, args...)
}

func (l *ContextLogger) Errortf(template string, args ...any) {
	Errortf(l.ctx, template, args...)
}

var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-csrf-token":        true,
	"x-xsrf-token":        true,
	"proxy-authorization": true,
}

func maskSensitiveHeaders(headers map[string]any) map[string]any {
	if len(headers) == 0 {
		return headers
	}
	result := make(map[string]any, len(headers))
	for k, v := range headers {
		if sensitiveHeaders[strings.ToLower(k)] {
			result[k] = "***"
		} else {
			result[k] = v
		}
	}
	return result
}
