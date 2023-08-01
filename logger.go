package gohera

import (
	"context"
	"fmt"

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
}

type (
	traceContext struct{}
)

// Entry 定义统一的日志写入方式
var logger *zap.Logger

type loggerConfig struct {
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
	Mode       string `json:"mode"` //环境
}

func initLoggerPool(config loggerConfig) {
	var allCore []zapcore.Core
	// 设置初始化字段
	filed := zap.Fields(
		zap.String("type", "go"),
		zap.String("project", GetAppName()),
	)
	core := zapcore.NewTee(allCore...)
	logger = zap.New(core, filed).WithOptions(zap.AddCaller())
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// NewTraceIDContext 创建跟踪ID上下文
func NewTraceIDContext(ctx context.Context, traceID any) context.Context {
	return context.WithValue(ctx, traceContext{}, traceID)
}

// FromTraceIDContext 从上下文中获取跟踪ID
func FromTraceContext(ctx context.Context) Trace {
	v := ctx.Value(traceContext{})
	if v != nil {
		if s, ok := v.(Trace); ok {
			return s
		}
	}
	return Trace{}
}

// StartSpan 开始一个追踪单元
func getContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		ctx = context.Background()
	}
	zapFiled := make([]zap.Field, 0)
	traceInfo := FromTraceContext(ctx)
	zapFiled = append(zapFiled, zap.String("trace_id", traceInfo.TraceId))
	zapFiled = append(zapFiled, zap.String("span_id", traceInfo.SpanId))
	zapFiled = append(zapFiled, zap.Int("user_id", traceInfo.UserId))
	zapFiled = append(zapFiled, zap.String("path", traceInfo.Path))
	zapFiled = append(zapFiled, zap.Int("status", traceInfo.Status))
	return zapFiled
}

// 判断其他类型--start
func StartSpan(ctx context.Context, format string, args ...interface{}) (string, []zap.Field) {
	//判断是否有context
	l := len(args)
	if l > 0 {
		if format == "" {
			return fmt.Sprint(args[:l-1]...), getContextFields(ctx)
		} else {
			return fmt.Sprintf(format, args[:l-1]...), getContextFields(ctx)
		}
	}
	return format, []zap.Field{}
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
