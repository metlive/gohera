package gohera

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/metlive/gohera/rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// 定义统一的日志写入方式
var logger *zap.Logger

type loggerConfig struct {
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
	Mode       string `json:"mode"` // 环境
}

// initLoggerPool 初始化日志连接池
// 根据配置初始化不同级别的日志输出（Debug, Info, Warn, Error）
func initLoggerPool(config loggerConfig) {
	// 调试级别
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	// 日志级别
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	// 警告级别
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	// 错误级别
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	
	// 文件输出 Core (保持 JSON 格式)
	cores := []zapcore.Core{
		getEncoderCore(fmt.Sprintf("./%s/server_debug.log", config.FilePath), debugPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_info.log", config.FilePath), infoPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_warn.log", config.FilePath), warnPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_error.log", config.FilePath), errorPriority, config),
	}

	// 新增：非正式环境添加控制台输出 (无格式纯文本)
	if config.Mode != "pro" {
		// 使用 zap.DebugLevel 允许输出 Debug 及以上所有级别日志
		cores = append(cores, getConsoleCore(zap.DebugLevel))
	}

	filed := zap.Fields(
		zap.String("x_type", "go"),
		zap.String("x_project", GetString("http.service")),
	)
	core := zapcore.NewTee(cores...)
	logger = zap.New(core, filed).WithOptions(zap.AddCallerSkip(1))
}

// getConsoleCore 获取控制台输出 Core (极简格式)
// 主要用于非正式环境下的控制台日志输出
func getConsoleCore(level zapcore.LevelEnabler) zapcore.Core {
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg", // 仅保留消息内容，留空其他 Key 以隐藏时间、级别等
	}), zapcore.AddSync(os.Stdout), level)

	// 返回一个 cleanConsoleCore 实例，用于精简控制台日志输出（忽略添加的字段，只输出 msg 字段的内容）
	return &cleanConsoleCore{Core: core}
}

type cleanConsoleCore struct {
	zapcore.Core
}

// With 实现 zapcore.Core 接口，添加字段
func (c *cleanConsoleCore) With(fields []zapcore.Field) zapcore.Core {
	return c // Ignore fields added via With
}

// Check 实现 zapcore.Core 接口，检查日志级别
func (c *cleanConsoleCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write 实现 zapcore.Core 接口，写入日志
// 忽略额外字段，仅输出日志消息
func (c *cleanConsoleCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(ent, nil) // Ignore fields during Write
}

// getEncoderCore 获取文件输出 Core 配置
// 负责配置日志文件的切割、格式（JSON）及输出级别
func getEncoderCore(fileName string, level zapcore.LevelEnabler, config loggerConfig) (core zapcore.Core) {
	// 每小时一个文件
	logf, _ := rotatelogs.New(fileName+"_%Y-%m-%d",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	
	// 修改处：不再包含 os.Stdout，只写入文件
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

// GetTraceContext 从 Context 中获取 Trace 信息
// 如果 Context 中不存在 Trace 信息，则返回空的 Trace 对象
func GetTraceContext(ctx context.Context) *Trace {
	if ctx == nil {
		return new(Trace)
	}
	// 尝试从标准 context.Value 获取 (兼容 Gin Context 和 request.Context())
	if v := ctx.Value(TraceCtx); v != nil {
		if t, ok := v.(*Trace); ok {
			return t
		}
	}
	return new(Trace)
}

// getContextFields 从 Context 中提取 Trace 信息并转换为 zap 字段
// 包含 trace_id, span_id, user_id, path, status 等信息
func getContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		ctx = context.Background()
	}
	zapFiled := make([]zap.Field, 0, 6)
	traceInfo := GetTraceContext(ctx)
	zapFiled = append(zapFiled, zap.String("x_trace_id", Ternary[string](traceInfo.TraceId == "", strings.ReplaceAll(uuid.NewString(), "-", ""), traceInfo.TraceId)))
	zapFiled = append(zapFiled, zap.String("x_span_id", Ternary[string](traceInfo.SpanId == "", SpanIdDefault, traceInfo.SpanId)))
	zapFiled = append(zapFiled, zap.Int("x_user_id", Ternary[int](traceInfo.UserId == 0, 0, traceInfo.UserId)))
	zapFiled = append(zapFiled, zap.String("x_path", traceInfo.Path))
	zapFiled = append(zapFiled, zap.Int("x_status", traceInfo.Status))
	zapFiled = append(zapFiled, zap.Any("x_header", traceInfo.Headers))
	return zapFiled
}

// StartSpan 处理日志格式化并从 Context 中提取跟踪信息
func StartSpan(ctx context.Context, format string, args ...any) (string, []zap.Field) {
	// 判断是否有context
	l := len(args)
	if l > 0 {
		return fmt.Sprintf(format, args[:l]...), getContextFields(ctx)
	}
	return format, getContextFields(ctx)
}

// Info 输出 Info 级别日志
func Info(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Info(str, filed...)
}

// Infotf 输出带格式化的 Info 级别日志
func Infotf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Info(str, filed...)
}

// Warn 输出 Warn 级别日志
func Warn(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Warn(str, filed...)
}

// Warntf 输出带格式化的 Warn 级别日志
func Warntf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Warn(str, filed...)
}

// Error 输出 Error 级别日志
func Error(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Error(str, filed...)
}

// Errortf 输出带格式化的 Error 级别日志
func Errortf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Error(str, filed...)
}

// ContextLogger 绑定了 Context 的日志记录器
type ContextLogger struct {
	ctx context.Context
}

// Ctx 创建一个绑定了 Context 的日志记录器
// 后续调用 Info/Warn/Error 等方法时无需再次传入 Context
func Ctx(ctx context.Context) *ContextLogger {
	return &ContextLogger{ctx: ctx}
}

// Info 输出 Info 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Info(args ...any) {
	Info(l.ctx, args...)
}

// Infotf 输出带格式化的 Info 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Infotf(template string, args ...any) {
	Infotf(l.ctx, template, args...)
}

// Warn 输出 Warn 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Warn(args ...any) {
	Warn(l.ctx, args...)
}

// Warntf 输出带格式化的 Warn 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Warntf(template string, args ...any) {
	Warntf(l.ctx, template, args...)
}

// Error 输出 Error 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Error(args ...any) {
	Error(l.ctx, args...)
}

// Errortf 输出带格式化的 Error 级别日志 (使用绑定的 Context)
func (l *ContextLogger) Errortf(template string, args ...any) {
	Errortf(l.ctx, template, args...)
}
