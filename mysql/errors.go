package mysql

import (
	"database/sql/driver"
	"errors"
	"io"
	"net"
	"slices"

	mysqldrv "github.com/go-sql-driver/mysql"
)

// MySQL 常见错误码,完整列表见
// https://dev.mysql.com/doc/mysql-errors/8.4/en/server-error-reference.html
// 与 https://dev.mysql.com/doc/mysql-errors/8.4/en/client-error-reference.html
const (
	// 服务器错误码
	ErrCodeTooManyConnections  = 1040 // ER_CON_COUNT_ERROR 连接数超限
	ErrCodeDBAccessDenied      = 1044 // ER_DBACCESS_DENIED_ERROR 无库权限
	ErrCodeAccessDenied        = 1045 // ER_ACCESS_DENIED_ERROR 账号密码错误
	ErrCodeNotNull             = 1048 // ER_BAD_NULL_ERROR 字段不允许为空
	ErrCodeUnknownDatabase     = 1049 // ER_BAD_DB_ERROR 数据库不存在
	ErrCodeTableExists         = 1050 // ER_TABLE_EXISTS_ERROR 表已存在
	ErrCodeUnknownTable        = 1051 // ER_BAD_TABLE_ERROR 表不存在(DDL)
	ErrCodeUnknownColumn       = 1054 // ER_BAD_FIELD_ERROR 字段不存在
	ErrCodeDuplicateEntry      = 1062 // ER_DUP_ENTRY 唯一约束冲突
	ErrCodeSyntaxError         = 1064 // ER_PARSE_ERROR SQL 语法错误
	ErrCodeNoSuchTable         = 1146 // ER_NO_SUCH_TABLE 表不存在(DML)
	ErrCodePacketTooLarge      = 1153 // ER_NET_PACKET_TOO_LARGE 超过 max_allowed_packet
	ErrCodeLockWaitTimeout     = 1205 // ER_LOCK_WAIT_TIMEOUT 锁等待超时
	ErrCodeDeadlock            = 1213 // ER_LOCK_DEADLOCK 死锁
	ErrCodeColumnOutOfRange    = 1264 // ER_WARN_DATA_OUT_OF_RANGE 字段数值超范围
	ErrCodeReadOnlyMode        = 1290 // ER_OPTION_PREVENTS_STATEMENT 服务器只读
	ErrCodeWrongValue          = 1292 // ER_TRUNCATED_WRONG_VALUE 值格式错误
	ErrCodeWrongValueForField  = 1366 // ER_TRUNCATED_WRONG_VALUE_FOR_FIELD 字段值格式错误
	ErrCodeDataTooLong         = 1406 // ER_DATA_TOO_LONG 数据超过字段长度
	ErrCodeFKParentRestricted  = 1451 // ER_ROW_IS_REFERENCED 外键:父行被引用
	ErrCodeFKChildMissing      = 1452 // ER_NO_REFERENCED_ROW 外键:子行引用缺失
	ErrCodeDataOutOfRange      = 1690 // ER_DATA_OUT_OF_RANGE 表达式数值超范围
	ErrCodeReadOnlyTransaction = 1792 // ER_CANT_EXECUTE_IN_READ_ONLY_TRANSACTION 只读事务
	ErrCodeReadOnly            = 1836 // ER_RUNNING_IN_READ_ONLY_MODE 只读模式

	// 客户端错误码(连接层,由驱动或代理返回)
	ErrCodeConnectionError     = 2002 // CR_CONNECTION_ERROR socket 连接失败
	ErrCodeConnectionHostError = 2003 // CR_CONN_HOST_ERROR 无法连接服务器
	ErrCodeServerGone          = 2006 // CR_SERVER_GONE_ERROR 服务器已断开
	ErrCodeServerLost          = 2013 // CR_SERVER_LOST 查询中连接丢失
	ErrCodeServerLostExtended  = 2055 // CR_SERVER_LOST_EXTENDED 连接丢失(扩展)
)

// Extract 从错误链中提取底层 *mysqldrv.MySQLError,可进一步读取 Number/SQLState/Message。
// err 不是 MySQL 错误时返回 ok=false。
func Extract(err error) (*mysqldrv.MySQLError, bool) {
	var me *mysqldrv.MySQLError
	if errors.As(err, &me) && me != nil {
		return me, true
	}
	return nil, false
}

// ErrorCode 返回 MySQL 错误码;err 不是 MySQL 错误时返回 0。
func ErrorCode(err error) uint16 {
	if me, ok := Extract(err); ok {
		return me.Number
	}
	return 0
}

// HasErrorCode 判断 err 是否为指定错误码之一的 MySQL 错误,支持被包装的错误。
func HasErrorCode(err error, codes ...uint16) bool {
	me, ok := Extract(err)
	return ok && slices.Contains(codes, me.Number)
}

// 约束与数据校验

// IsDuplicateEntry 判断错误是否为唯一约束冲突(ER_DUP_ENTRY,1062)。
func IsDuplicateEntry(err error) bool {
	return HasErrorCode(err, ErrCodeDuplicateEntry)
}

// IsForeignKeyViolation 判断错误是否为外键约束冲突
// (1452 新增/更新的子行缺少父行,1451 删除/更新的父行被子行引用)。
func IsForeignKeyViolation(err error) bool {
	return HasErrorCode(err, ErrCodeFKChildMissing, ErrCodeFKParentRestricted)
}

// IsNotNullViolation 判断错误是否为非空约束冲突(ER_BAD_NULL_ERROR,1048)。
func IsNotNullViolation(err error) bool {
	return HasErrorCode(err, ErrCodeNotNull)
}

// IsDataTooLong 判断错误是否为数据超过字段长度(ER_DATA_TOO_LONG,1406)。
func IsDataTooLong(err error) bool {
	return HasErrorCode(err, ErrCodeDataTooLong)
}

// IsOutOfRange 判断错误是否为数值超出范围(字段 1264 或表达式 1690)。
func IsOutOfRange(err error) bool {
	return HasErrorCode(err, ErrCodeColumnOutOfRange, ErrCodeDataOutOfRange)
}

// IsTruncatedWrongValue 判断错误是否为值格式错误(日期/时间/数字/枚举,1292/1366)。
func IsTruncatedWrongValue(err error) bool {
	return HasErrorCode(err, ErrCodeWrongValue, ErrCodeWrongValueForField)
}

// 锁与事务

// IsDeadlock 判断错误是否为死锁(ER_LOCK_DEADLOCK,1213),此时整个事务已被回滚。
func IsDeadlock(err error) bool {
	return HasErrorCode(err, ErrCodeDeadlock)
}

// IsLockWaitTimeout 判断错误是否为锁等待超时(ER_LOCK_WAIT_TIMEOUT,1205),
// InnoDB 默认只回滚当前语句,事务未提交部分需自行处理。
func IsLockWaitTimeout(err error) bool {
	return HasErrorCode(err, ErrCodeLockWaitTimeout)
}

// IsReadOnly 判断错误是否因服务器或事务只读(1290/1792/1836)。
func IsReadOnly(err error) bool {
	return HasErrorCode(err, ErrCodeReadOnlyMode, ErrCodeReadOnlyTransaction, ErrCodeReadOnly)
}

// 连接与服务

// IsConnectionError 判断错误是否为连接类错误(无法建立连接或连接已断开)。
// 覆盖:客户端错误码(2002/2003/2006/2013/2055)、database/sql 的 bad connection、
// 驱动的 invalid connection/malformed packet、底层网络错误(dial/read/write 失败)。
// 不包含 context 取消或超时。
func IsConnectionError(err error) bool {
	if HasErrorCode(err, ErrCodeConnectionError, ErrCodeConnectionHostError,
		ErrCodeServerGone, ErrCodeServerLost, ErrCodeServerLostExtended) {
		return true
	}
	if errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, mysqldrv.ErrInvalidConn) ||
		errors.Is(err, mysqldrv.ErrMalformPkt) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}

// IsTooManyConnections 判断错误是否为服务器连接数超限(ER_CON_COUNT_ERROR,1040)。
func IsTooManyConnections(err error) bool {
	return HasErrorCode(err, ErrCodeTooManyConnections)
}

// 权限与对象

// IsAccessDenied 判断错误是否为认证失败或无权限(1044/1045)。
func IsAccessDenied(err error) bool {
	return HasErrorCode(err, ErrCodeAccessDenied, ErrCodeDBAccessDenied)
}

// IsUnknownDatabase 判断错误是否为数据库不存在(ER_BAD_DB_ERROR,1049)。
func IsUnknownDatabase(err error) bool {
	return HasErrorCode(err, ErrCodeUnknownDatabase)
}

// IsNoSuchTable 判断错误是否为表不存在(1051 DDL 场景 / 1146 DML 场景)。
func IsNoSuchTable(err error) bool {
	return HasErrorCode(err, ErrCodeUnknownTable, ErrCodeNoSuchTable)
}

// IsTableExists 判断错误是否为建表时表已存在(ER_TABLE_EXISTS_ERROR,1050)。
func IsTableExists(err error) bool {
	return HasErrorCode(err, ErrCodeTableExists)
}

// IsUnknownColumn 判断错误是否为字段不存在(ER_BAD_FIELD_ERROR,1054)。
func IsUnknownColumn(err error) bool {
	return HasErrorCode(err, ErrCodeUnknownColumn)
}

// IsSyntaxError 判断错误是否为 SQL 语法错误(ER_PARSE_ERROR,1064)。
func IsSyntaxError(err error) bool {
	return HasErrorCode(err, ErrCodeSyntaxError)
}

// IsPacketTooLarge 判断错误是否超过 max_allowed_packet(服务端 1153 或驱动侧限制)。
func IsPacketTooLarge(err error) bool {
	return HasErrorCode(err, ErrCodePacketTooLarge) || errors.Is(err, mysqldrv.ErrPktTooLarge)
}

// 重试

// IsRetryable 判断错误是否通常可通过重试解决:死锁、锁等待超时、连接类错误。
// 重试前需保证操作可安全重放(建议整事务重试并配合退避)。
func IsRetryable(err error) bool {
	return IsDeadlock(err) || IsLockWaitTimeout(err) || IsConnectionError(err)
}
