package rotatelogs

import (
	"time"
)

const (
	optkeyClock         = "clock"
	optkeyHandler       = "handler"
	optkeyLinkName      = "link-name"
	optkeyMaxAge        = "max-age"
	optkeyRotationTime  = "rotation-time"
	optkeyRotationSize  = "rotation-size"
	optkeyRotationCount = "rotation-count"
	optkeyForceNewFile  = "force-new-file"
)

type Option interface {
	Name() string
	Value() any
}

type option struct {
	name  string
	value any
}

func SelectorNew(name string, value any) *option {
	return &option{
		name:  name,
		value: value,
	}
}

func (o *option) Name() string {
	return o.name
}
func (o *option) Value() any {
	return o.value
}

// WithClock 设置时钟对象
// RotateLogs 将使用该时钟来确定当前时间
// 默认使用 rotatelogs.Local (本地时间)
func WithClock(c Clock) Option {
	return SelectorNew(optkeyClock, c)
}

// WithLocation creates a new Option that sets up a
// "Clock" interface that the RotateLogs object will use
// to determine the current time.
//
// This optin works by always returning the in the given
// location.
func WithLocation(loc *time.Location) Option {
	return SelectorNew(optkeyClock, clockFn(func() time.Time {
		return time.Now().In(loc)
	}))
}

// WithLinkName 设置软链接名称
// 会创建一个指向最新日志文件的软链接
func WithLinkName(s string) Option {
	return SelectorNew(optkeyLinkName, s)
}

// WithMaxAge 设置日志最大保存时间
// 超过该时间的日志文件将被清理
func WithMaxAge(d time.Duration) Option {
	return SelectorNew(optkeyMaxAge, d)
}

// WithRotationTime 设置日志轮转间隔
func WithRotationTime(d time.Duration) Option {
	return SelectorNew(optkeyRotationTime, d)
}

// WithRotationSize 设置日志轮转大小
// 当文件大小超过该值时进行轮转
func WithRotationSize(s int64) Option {
	return SelectorNew(optkeyRotationSize, s)
}

// WithRotationCount 设置保留的日志文件数量
func WithRotationCount(n uint) Option {
	return SelectorNew(optkeyRotationCount, n)
}

// WithHandler 设置事件处理器
// 目前支持 `FileRotated` 事件
func WithHandler(h Handler) Option {
	return SelectorNew(optkeyHandler, h)
}

// ForceNewFile 强制创建新文件
// 确保每次调用 New() 时都创建新文件，如果文件名已存在则进行轮转
func ForceNewFile() Option {
	return SelectorNew(optkeyForceNewFile, true)
}
