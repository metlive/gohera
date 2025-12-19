package mysql

import (
	"context"

	"xorm.io/xorm"
)

// Session 会话包装类
type Session struct {
	*xorm.Session
}

// Tx 事务包装类
type Tx struct {
	*Session
}

// NewSession 创建新会话
func (db *DB) NewSession() *Session {
	return &Session{db.Engine.NewSession()}
}

// Session 获取会话实例
func (db *DB) Session() *Session {
	return db.NewSession()
}

// Begin 开始事务
func (db *DB) Begin() (*Tx, error) {
	return db.BeginTx(context.Background())
}

// BeginTx 开始带上下文的事务
func (db *DB) BeginTx(ctx context.Context) (*Tx, error) {
	session := db.NewSession()
	session.Session = session.Session.Context(ctx)
	err := session.Begin()
	if err != nil {
		session.Close()
		return nil, err
	}
	return &Tx{session}, nil
}

// WithTransaction 在事务中执行函数
func (db *DB) WithTransaction(fn func(*Tx) error) error {
	return db.WithTransactionCtx(context.Background(), fn)
}

// WithTransactionCtx 在带上下文的事务中执行函数
func (db *DB) WithTransactionCtx(ctx context.Context, fn func(*Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Commit 提交事务
func (tx *Tx) Commit() error {
	err := tx.Session.Commit()
	tx.Session.Close()
	return err
}

// Rollback 回滚事务
func (tx *Tx) Rollback() error {
	err := tx.Session.Rollback()
	tx.Session.Close()
	return err
}

// Context 返回带有上下文的 Session
func (s *Session) Context(ctx context.Context) *Session {
	s.Session = s.Session.Context(ctx)
	return s
}

// Table 指定表名
func (s *Session) Table(tableNameOrStruct interface{}) *Session {
	s.Session = s.Session.Table(tableNameOrStruct)
	return s
}

// SQL 执行自定义 SQL
func (s *Session) SQL(query interface{}, args ...interface{}) *Session {
	s.Session = s.Session.SQL(query, args...)
	return s
}

// Where 添加查询条件
func (s *Session) Where(query interface{}, args ...interface{}) *Session {
	s.Session = s.Session.Where(query, args...)
	return s
}

// Limit 分页查询
func (s *Session) Limit(limit int, start ...int) *Session {
	s.Session = s.Session.Limit(limit, start...)
	return s
}

// Desc 降序
func (s *Session) Desc(colNames ...string) *Session {
	s.Session = s.Session.Desc(colNames...)
	return s
}

// Asc 升序
func (s *Session) Asc(colNames ...string) *Session {
	s.Session = s.Session.Asc(colNames...)
	return s
}

// ID 指定主键
func (s *Session) ID(id interface{}) *Session {
	s.Session = s.Session.ID(id)
	return s
}

// In 指定 IN 条件
func (s *Session) In(column string, args ...interface{}) *Session {
	s.Session = s.Session.In(column, args...)
	return s
}

// FindAndCount 查询列表并返回总数
func (s *Session) FindAndCount(rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
	return s.Session.FindAndCount(rowsSlicePtr, condiBean...)
}

// Get 查询单条记录
func (s *Session) Get(bean interface{}) (bool, error) {
	return s.Session.Get(bean)
}

// Insert 插入记录
func (s *Session) Insert(beans ...interface{}) (int64, error) {
	return s.Session.Insert(beans...)
}

// Update 更新记录
func (s *Session) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
	return s.Session.Update(bean, condiBean...)
}

// Delete 删除记录
func (s *Session) Delete(bean interface{}) (int64, error) {
	return s.Session.Delete(bean)
}
