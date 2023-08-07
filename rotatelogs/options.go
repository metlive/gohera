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

// WithClock creates a new Option that sets a clock
// that the RotateLogs object will use to determine
// the current time.
//
// By default rotatelogs.Local, which returns the
// current time in the local time zone, is used. If you
// would rather use UTC, use rotatelogs.UTC as the argument
// to this option, and pass it to the constructor.
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

// WithLinkName creates a new Option that sets the
// symbolic link name that gets linked to the current
// file name being used.
func WithLinkName(s string) Option {
	return SelectorNew(optkeyLinkName, s)
}

// WithMaxAge creates a new Option that sets the
// max age of a log file before it gets purged from
// the file system.
func WithMaxAge(d time.Duration) Option {
	return SelectorNew(optkeyMaxAge, d)
}

// WithRotationTime creates a new Option that sets the
// time between rotation.
func WithRotationTime(d time.Duration) Option {
	return SelectorNew(optkeyRotationTime, d)
}

// WithRotationSize creates a new Option that sets the
// log file size between rotation.
func WithRotationSize(s int64) Option {
	return SelectorNew(optkeyRotationSize, s)
}

// WithRotationCount creates a new Option that sets the
// number of files should be kept before it gets
// purged from the file system.
func WithRotationCount(n uint) Option {
	return SelectorNew(optkeyRotationCount, n)
}

// WithHandler creates a new Option that specifies the
// Handler object that gets invoked when an event occurs.
// Currently `FileRotated` event is supported
func WithHandler(h Handler) Option {
	return SelectorNew(optkeyHandler, h)
}

// ForceNewFile ensures a new file is created every time New()
// is called. If the base file name already exists, an implicit
// rotation is performed
func ForceNewFile() Option {
	return SelectorNew(optkeyForceNewFile, true)
}
