package session

import (
    "database/sql"
    "geeorm/log"
    "strings"
)

// Session keep a pointer to sql.DB and provides all execution of all
// kind of database operations.
type Session struct {
	// 使用 sql.Open() 方法连接数据库成功之后返回的指针。
    db      *sql.DB

    // 用于拼接 SQL 语句和 SQL 语句中占位符的对应值，调用 Raw 方法改变这两个值。
    sql     strings.Builder
    sqlVars []interface{}
}

// 封装原生 SQL 执行方法：统一打印日志、执行完成后清空 (s *Session).sql 和 (s *Session).sqlVars，使得 Session 可在多个 SQL 之间复用。

// New creates a instance of Session
func New(db *sql.DB) *Session {
    return &Session{db: db}
}

// Clear initialize the state of a session
func (s *Session) Clear() {
    s.sql.Reset()
    s.sqlVars = nil
}

// DB returns *sql.DB
func (s *Session) DB() *sql.DB {
    return s.db
}

// Exec raw sql with sqlVars
func (s *Session) Exec() (result sql.Result, err error) {
    defer s.Clear()
    log.Info(s.sql.String(), s.sqlVars)

    // 执行 SQL 语句：如果是查询语句，不会返回相关的记录。
    if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
        log.Error(err)
    }
    return
}

// QueryRow gets a record from db
func (s *Session) QueryRow() *sql.Row {
    defer s.Clear()
    log.Info(s.sql.String(), s.sqlVars)
    // 执行查询 SQL 语句，传入 SQL 语句和占位符的值。
    return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows gets a list of records from db
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
    defer s.Clear()
    log.Info(s.sql.String(), s.sqlVars)
    if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
        log.Error(err)
    }
    return
}

// Raw appends sql and sqlVars
func (s *Session) Raw(sql string, values ...interface{}) *Session {
    s.sql.WriteString(sql)
    s.sql.WriteString(" ")
    s.sqlVars = append(s.sqlVars, values...)
    return s
}