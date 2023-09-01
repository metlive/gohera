package gohera

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/metlive/gohera/rotatelogs"
	"os"
	"strings"
	"time"

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
	Mode       string `json:"mode"` //环境
}

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
	cores := []zapcore.Core{
		getEncoderCore(fmt.Sprintf("./%s/server_debug.log", config.FilePath), debugPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_info.log", config.FilePath), infoPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_warn.log", config.FilePath), warnPriority, config),
		getEncoderCore(fmt.Sprintf("./%s/server_error.log", config.FilePath), errorPriority, config),
	}
	// 设置初始化字段
	filed := zap.Fields(
		zap.String("type", "go"),
		zap.String("project", GetString("http.service")),
	)
	core := zapcore.NewTee(cores...)
	logger = zap.New(core, filed).WithOptions(zap.AddCallerSkip(1))
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore(fileName string, level zapcore.LevelEnabler, config loggerConfig) (core zapcore.Core) {
	// 每小时一个文件
	logf, _ := rotatelogs.New(fileName+"_%Y-%m-%d",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	var writer zapcore.WriteSyncer
	if config.Mode != "pro" {
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(logf))
	} else {
		writer = zapcore.AddSync(logf)
	}
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

// 从上下文中获取跟踪ID
func GetTraceContext(ctx context.Context) *Trace {
	if ctxValue, ok := ctx.(*gin.Context); ok {
		return ctxValue.MustGet(TraceCtx).(*Trace)
	} else {
		trace := ctx.Value(TraceCtx)
		if trace != nil {
			return trace.(*Trace)
		}
	}
	return new(Trace)
}

// 开始一个追踪单元
func getContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		ctx = context.Background()
	}
	zapFiled := make([]zap.Field, 0)
	traceInfo := GetTraceContext(ctx)
	zapFiled = append(zapFiled, zap.String("x_trace_id", Ternary[string](traceInfo.TraceId == "", strings.ReplaceAll(uuid.NewString(), "-", ""), traceInfo.TraceId)))
	zapFiled = append(zapFiled, zap.String("x_span_id", Ternary[string](traceInfo.SpanId == "", SpanIdDefault, traceInfo.SpanId)))
	zapFiled = append(zapFiled, zap.Int("x_user_id", Ternary[int](traceInfo.UserId == 0, 0, traceInfo.UserId)))
	zapFiled = append(zapFiled, zap.String("x_path", traceInfo.Path))
	zapFiled = append(zapFiled, zap.Int("x_status", traceInfo.Status))
	zapFiled = append(zapFiled, zap.Any("x_header", traceInfo.Headers))
	return zapFiled
}

// 判断其他类型--start
func StartSpan(ctx context.Context, format string, args ...interface{}) (string, []zap.Field) {
	//判断是否有context
	l := len(args)
	if l > 0 {
		return fmt.Sprintf(format, args[:l]...), getContextFields(ctx)
	}
	return format, getContextFields(ctx)
}

func Info(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Info(str, filed...)
}

func Infotf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Info(str, filed...)
}

func Warn(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Warn(str, filed...)
}

func Warntf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Warn(str, filed...)
}

func Error(ctx context.Context, args ...any) {
	str, filed := StartSpan(ctx, "%v", args...)
	logger.Error(str, filed...)
}

func Errortf(ctx context.Context, template string, args ...any) {
	str, filed := StartSpan(ctx, template, args...)
	logger.Error(str, filed...)
}
