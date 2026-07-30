package strftime

type Option interface {
	Name() string
	Value() any
}

type option struct {
	name  string
	value any
}

func (o *option) Name() string { return o.name }
func (o *option) Value() any   { return o.value }

const optSpecificationSet = `opt-specification-set`

// WithSpecificationSet 允许指定自定义规范集
func WithSpecificationSet(ds SpecificationSet) Option {
	return &option{
		name:  optSpecificationSet,
		value: ds,
	}
}

type optSpecificationPair struct {
	name     byte
	appender Appender
}

const optSpecification = `opt-specification`

// WithSpecification 允许动态创建新的规范
func WithSpecification(b byte, a Appender) Option {
	return &option{
		name: optSpecification,
		value: &optSpecificationPair{
			name:     b,
			appender: a,
		},
	}
}

// WithMilliseconds 指定 %b 模式（b为参数）解释为毫秒（3位，零填充）
func WithMilliseconds(b byte) Option {
	return WithSpecification(b, Milliseconds())
}

// WithMicroseconds 指定 %b 模式（b为参数）解释为微秒（3位，零填充）
func WithMicroseconds(b byte) Option {
	return WithSpecification(b, Microseconds())
}

// WithUnixSeconds 指定 %b 模式（b为参数）解释为 Unix 时间戳（秒）
func WithUnixSeconds(b byte) Option {
	return WithSpecification(b, UnixSeconds())
}
