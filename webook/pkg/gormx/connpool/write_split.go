package connpool

import (
	"context"
	"database/sql"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

type WriteSplit struct {
	master gorm.ConnPool
	slaves []gorm.ConnPool
	pos    atomicx.Value[int]
}

func (w *WriteSplit) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.master.(gorm.TxBeginner).BeginTx(ctx, opts)
}

func (w *WriteSplit) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// 可以默认返回master
	return w.master.PrepareContext(ctx, query)
}

func (w *WriteSplit) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.master.ExecContext(ctx, query, args...)
}

func (w *WriteSplit) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// slaves 要考虑 负载均衡，搞个轮询
	pos := w.pos.Load()
	slave := w.slaves[pos]
	w.pos.Store((pos + 1) % len(w.slaves))
	return slave.QueryContext(ctx, query, args...)
}

func (w *WriteSplit) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// slaves 要考虑 负载均衡，搞个轮询
	pos := w.pos.Load()
	slave := w.slaves[pos]
	w.pos.Store((pos + 1) % len(w.slaves))
	return slave.QueryRowContext(ctx, query, args...)
}
