package mysql

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	mysqldrv "github.com/go-sql-driver/mysql"
)

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}
	if err := cfg.validateAndNormalize(); err == nil {
		t.Fatal("expected error")
	}
	cfg = &Config{Host: "127.0.0.1", Database: "db", User: "u"}
	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3306 {
		t.Fatalf("port=%d", cfg.Port)
	}
	if cfg.MaxIdleConns != DefaultMaxIdleConns {
		t.Fatalf("idle=%d", cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns != DefaultMaxOpenConns {
		t.Fatalf("open=%d", cfg.MaxOpenConns)
	}
	if cfg.MaxLifeTime != DefaultMaxLifeTime {
		t.Fatalf("life=%v", cfg.MaxLifeTime)
	}
}

func TestResolveShowSQL(t *testing.T) {
	cfg := &Config{Env: "dev"}
	if !cfg.resolveShowSQL() {
		t.Fatal("dev should show sql")
	}
	cfg = &Config{Env: "prod", ShowSQL: Bool(true)}
	if !cfg.resolveShowSQL() {
		t.Fatal("explicit true")
	}
	cfg = &Config{Env: "dev", ShowSQL: Bool(false)}
	if cfg.resolveShowSQL() {
		t.Fatal("explicit false")
	}
}

func TestInitOnceCompat(t *testing.T) {
	p := InitOnce(&Config{Host: "h", Database: "d", User: "u", MaxLifeTime: time.Second})
	if p == nil || p.config == nil {
		t.Fatal("nil pool")
	}
}

func TestNewNilConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestErrorHelpers(t *testing.T) {
	mysqlErr := func(code uint16) error {
		return &mysqldrv.MySQLError{Number: code, Message: "test"}
	}

	tests := []struct {
		name string
		err  error
		fn   func(error) bool
	}{
		{"duplicate entry", mysqlErr(ErrCodeDuplicateEntry), IsDuplicateEntry},
		{"duplicate entry wrapped", fmt.Errorf("xorm: %w", mysqlErr(ErrCodeDuplicateEntry)), IsDuplicateEntry},
		{"fk child missing", mysqlErr(ErrCodeFKChildMissing), IsForeignKeyViolation},
		{"fk parent restricted", mysqlErr(ErrCodeFKParentRestricted), IsForeignKeyViolation},
		{"not null", mysqlErr(ErrCodeNotNull), IsNotNullViolation},
		{"data too long", mysqlErr(ErrCodeDataTooLong), IsDataTooLong},
		{"column out of range", mysqlErr(ErrCodeColumnOutOfRange), IsOutOfRange},
		{"data out of range", mysqlErr(ErrCodeDataOutOfRange), IsOutOfRange},
		{"wrong value", mysqlErr(ErrCodeWrongValue), IsTruncatedWrongValue},
		{"wrong value for field", mysqlErr(ErrCodeWrongValueForField), IsTruncatedWrongValue},
		{"deadlock", mysqlErr(ErrCodeDeadlock), IsDeadlock},
		{"lock wait timeout", mysqlErr(ErrCodeLockWaitTimeout), IsLockWaitTimeout},
		{"read only mode", mysqlErr(ErrCodeReadOnlyMode), IsReadOnly},
		{"read only transaction", mysqlErr(ErrCodeReadOnlyTransaction), IsReadOnly},
		{"too many connections", mysqlErr(ErrCodeTooManyConnections), IsTooManyConnections},
		{"access denied", mysqlErr(ErrCodeAccessDenied), IsAccessDenied},
		{"db access denied", mysqlErr(ErrCodeDBAccessDenied), IsAccessDenied},
		{"unknown database", mysqlErr(ErrCodeUnknownDatabase), IsUnknownDatabase},
		{"no such table", mysqlErr(ErrCodeNoSuchTable), IsNoSuchTable},
		{"unknown table", mysqlErr(ErrCodeUnknownTable), IsNoSuchTable},
		{"table exists", mysqlErr(ErrCodeTableExists), IsTableExists},
		{"unknown column", mysqlErr(ErrCodeUnknownColumn), IsUnknownColumn},
		{"syntax error", mysqlErr(ErrCodeSyntaxError), IsSyntaxError},
		{"packet too large server", mysqlErr(ErrCodePacketTooLarge), IsPacketTooLarge},
		{"packet too large client", mysqldrv.ErrPktTooLarge, IsPacketTooLarge},
		{"conn server lost", mysqlErr(ErrCodeServerLost), IsConnectionError},
		{"conn bad conn", fmt.Errorf("query: %w", driver.ErrBadConn), IsConnectionError},
		{"conn invalid conn", mysqldrv.ErrInvalidConn, IsConnectionError},
		{"conn net op", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}, IsConnectionError},
		{"retryable deadlock", fmt.Errorf("tx: %w", mysqlErr(ErrCodeDeadlock)), IsRetryable},
		{"retryable bad conn", driver.ErrBadConn, IsRetryable},
	}
	for _, tt := range tests {
		if !tt.fn(tt.err) {
			t.Errorf("%s: expected true, got false", tt.name)
		}
	}

	// 非 MySQL 错误与 nil 一律不命中
	for _, fn := range []func(error) bool{
		IsDuplicateEntry, IsForeignKeyViolation, IsNotNullViolation, IsDataTooLong,
		IsOutOfRange, IsTruncatedWrongValue, IsDeadlock, IsLockWaitTimeout, IsReadOnly,
		IsConnectionError, IsTooManyConnections, IsAccessDenied, IsUnknownDatabase,
		IsNoSuchTable, IsTableExists, IsUnknownColumn, IsSyntaxError, IsPacketTooLarge,
		IsRetryable,
	} {
		if fn(errors.New("some error")) {
			t.Errorf("plain error: unexpected true")
		}
		if fn(nil) {
			t.Errorf("nil error: unexpected true")
		}
	}
	// context 取消不算连接错误
	if IsConnectionError(context.Canceled) || IsRetryable(context.Canceled) {
		t.Error("context.Canceled should not be connection/retryable error")
	}

	if got := ErrorCode(errors.New("some error")); got != 0 {
		t.Errorf("ErrorCode(plain) = %d, want 0", got)
	}
	if got := ErrorCode(mysqlErr(ErrCodeDuplicateEntry)); got != ErrCodeDuplicateEntry {
		t.Errorf("ErrorCode() = %d, want %d", got, ErrCodeDuplicateEntry)
	}
	if !HasErrorCode(mysqlErr(ErrCodeDeadlock), ErrCodeDeadlock, ErrCodeDuplicateEntry) {
		t.Error("HasErrorCode should match")
	}
	if HasErrorCode(errors.New("some error"), ErrCodeDeadlock) {
		t.Error("HasErrorCode(plain) should be false")
	}
	me, ok := Extract(fmt.Errorf("outer: %w", mysqlErr(ErrCodeSyntaxError)))
	if !ok || me == nil || me.Number != ErrCodeSyntaxError {
		t.Errorf("Extract: ok=%v number=%v", ok, me)
	}
	if _, ok := Extract(errors.New("some error")); ok {
		t.Error("Extract(plain) should be false")
	}
}
