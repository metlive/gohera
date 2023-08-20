package gohera

const (
	Success         = 0
	ErrSystem       = 1000000 // 系统错误
	ErrUnknown      = 9999999 // 未知错误
	ErrInternal     = 1010101 // 内部错误
	ErrMysql        = 1010102 // Mysql错误
	ErrRedis        = 1010103 // Redis错误
	ErrAccessToken  = 1010201 // token错误
	ErrParam        = 1010301 // 参数错误
	DefaultErrorMsg = 1000001
)

const (
	TraceId        = "x-trace-id"
	SpanId         = "x-span-id"
	UserId         = "x-user-id"
	TraceCtx       = "trace-ctx"
	TraceHeaderCtx = "trace-header-ctx"
	SpanIdDefault  = "0"
)
