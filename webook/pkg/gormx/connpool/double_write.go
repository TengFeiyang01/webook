package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

const (
	patternSrcOnly  = "patternSrcOnly"
	patternSrcFirst = "patternSrcFirst"
	patternDstOnly  = "patternDstOnly"
	patternDstFirst = "DST_ONLY"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case patternSrcOnly:
		tx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			src:     tx,
			pattern: pattern,
		}, err
	case patternSrcFirst:
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志 不做处理
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil
	case patternDstOnly:
		tx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			dst:     tx,
			pattern: pattern,
		}, err
	case patternDstFirst:
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志 不做处理
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil
	default:
		return nil, errUnknownPattern
	}
}

// PrepareContext 实现了 sql.PrepareContext 接口
// Prepare 的语句会进来这里
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("双写模式不支持 Prepare 语句")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	pattern := d.pattern.Load()
	// 任何非查询语句会进来这里
	switch pattern {
	case patternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		// 先写源库，再写目标库
		// 1. 先写源库
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		// 2. 再写目标库
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, err
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case patternDstFirst:
		// 先写目标库，再写源库
		// 1. 先写目标库
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		// 2. 再写源库
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	pattern := d.pattern.Load()

	switch pattern {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		// 这里怎么构造这个 sql.Rows 呢？
		panic(errUnknownPattern)
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这里怎么构造这个 sql.Rows 呢？
		panic(errUnknownPattern)
	}
}

type DoubleWritePoolTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
}

func (d *DoubleWritePoolTx) Commit() error {
	switch d.pattern {
	case patternSrcOnly:
		return d.src.Commit()
	case patternSrcFirst:
		// commit 失败了，怎么办
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Commit()
	case patternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.pattern {
	case patternSrcOnly:
		return d.src.Rollback()
	case patternSrcFirst:
		// rollback 失败了，怎么办
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Rollback()
	case patternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// 记录日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("implement me")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 任何非查询语句会进来这里
	switch d.pattern {
	case patternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		// 先写源库，再写目标库
		// 1. 先写源库
		if d.src == nil {
			return nil, nil
		}
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.dst == nil {
			return res, nil
		}
		// 2. 再写目标库
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, err
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case patternDstFirst:
		// 先写目标库，再写源库
		// 1. 先写目标库
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.src == nil {
			return res, nil
		}
		// 2. 再写源库
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		// 这里怎么构造这个 sql.Rows 呢？
		panic(errUnknownPattern)
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这里怎么构造这个 sql.Rows 呢？
		panic(errUnknownPattern)
	}
}
