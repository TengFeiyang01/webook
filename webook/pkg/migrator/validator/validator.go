package validator

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator/events"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	baseValidator
	batchSize int
	utime     int64
	// 如果没有数据了，就睡眠
	// 如果不是正数，那么就说明直接返回，结束这一次的循环
	// 我很厌恶这种特殊值有特殊含义的做法，但是不得不搞
	sleepInterval time.Duration
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.LoggerV1,
	producer events.Producer,
) *Validator[T] {
	return &Validator[T]{
		baseValidator: baseValidator{
			base:      base,
			target:    target,
			direction: direction,
			l:         l,
			producer:  producer,
		},
		batchSize: 100,
		// 默认是全量校验，并且数据没了就结束
		sleepInterval: 0,
	}
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

// Validate 执行校验。
// 分成两步：
// 1. from => to
func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.baseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.targetToBase(ctx)
	})
	return eg.Wait()
}

// baseToTarget 从 first 到 second 的验证
func (v *Validator[T]) baseToTarget(ctx context.Context) error {
	offset := 0
	for {
		var src T
		// 这里假定主键的规范都是叫做 id，基本上大部分公司都有这种规范
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.base.WithContext(dbCtx).
			Order("id").
			Where("utime >= ?", v.utime).
			Offset(offset).First(&src).Error
		cancel()
		switch err {
		case gorm.ErrRecordNotFound:
			// 已经没有数据了
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		case context.Canceled, context.DeadlineExceeded:
			// 退出循环
			return nil
		case nil:
			v.dstDiff(ctx, src)
		default:
			v.l.Error("src => dst 查询源表失败", logger.Error(err))
		}
		offset++
	}
}

func (v *Validator[T]) dstDiff(ctx context.Context, src T) {
	var dst T
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.target.WithContext(dbCtx).
		Where("id=?", src.ID()).First(&dst).Error
	cancel()
	// 这边要考虑不同的 error
	switch err {
	case gorm.ErrRecordNotFound:
		v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
	case nil:
		// 查询到了数据
		equal := src.CompareTo(dst)
		if !equal {
			v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
		}
	default:
		v.l.Error("src => dst 查询目标表失败", logger.Error(err))
	}
}

// targetToBase 反过来，执行 target 到 base 的验证
// 这是为了找出 dst 中多余的数据
func (v *Validator[T]) targetToBase(ctx context.Context) error {
	// 这个我们只需要找出 src 中不存在的 id 就可以了
	offset := 0
	for {
		var ts []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.target.WithContext(dbCtx).Model(new(T)).Select("id").Offset(offset).
			Limit(v.batchSize).Find(&ts).Error
		cancel()
		switch err {
		case gorm.ErrRecordNotFound:
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				// 在 sleep 的时候。不需要调整偏移量
				continue
			}
		case context.DeadlineExceeded, context.Canceled:
			return nil
		case nil:
			v.srcMissingRecords(ctx, ts)
		default:
			v.l.Error("dst => src 查询目标表失败", logger.Error(err))
		}
		if len(ts) < v.batchSize {
			// 数据没了
			return nil
		}
		offset += v.batchSize
	}
}

func (v *Validator[T]) srcMissingRecords(ctx context.Context, ts []T) {
	ids := slice.Map(ts, func(idx int, src T) int64 {
		return src.ID()
	})
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	base := v.base.WithContext(dbCtx)
	var srcTs []T
	err := base.Select("id").Where("id IN ?", ids).Find(&srcTs).Error
	switch err {
	case gorm.ErrRecordNotFound:
		// 说明 ids 全部没有
		v.notifySrcMissing(ts)
	case nil:
		// 计算差集
		missing := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifySrcMissing(missing)
	default:
		v.l.Error("dst => src 查询源表失败", logger.Error(err))
	}
}

func (v *Validator[T]) notifySrcMissing(ts []T) {
	for _, t := range ts {
		v.notify(t.ID(), events.InconsistentEventTypeBaseMissing)
	}
}
