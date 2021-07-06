package geeorm

import (
	"database/sql"
	"geeorm/log"
	"geeorm/session"
)

// Engine is the main struct of geeorm, manages all db sessions and transactions.
type Engine struct {
	db *sql.DB
}

// NewEngine create a instance of Engine
// connect database and ping it to test whether it's alive
func NewEngine(driver, source string) (e *Engine, err error) {

	// 连接数据库：输入驱动名称、数据库名称，返回 sql.DB 实例指针。
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	// 测试数据库连接是否正常可用。
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	e = &Engine{db: db}
	log.Info("Connect database success")
	return
}

// Close database connection
// 关闭数据库连接。
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// NewSession creates a new session for next operations
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}