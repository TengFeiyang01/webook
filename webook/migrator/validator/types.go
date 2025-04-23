package validator

import (
	"errors"
	"github.com/TengFeiyang01/webook/webook/migrator"
	"github.com/TengFeiyang01/webook/webook/migrator/events"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB // 以什么为准
	target *gorm.DB // 校验的是谁的数据

	l logger.LoggerV1

	direction string
	p         events.Producer
	batchSize int

	highLoad *atomicx.Value[bool]

	utime int64
	// <=0 说明 直接退出校验循环
	// >0 真的是 sleep
	sleepInterval time.Duration
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1, direction string, p events.Producer) *Validator[T] {
	highLoad := atomicx.NewValueOf(false)
	go func() {
		// 去查询数据库的状态
		// 结合本地的 CPU 内存判断
	}()
	return &Validator[T]{base: base, target: target, l: l, direction: direction, p: p,
		highLoad: highLoad,
	}
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
// 全量校验，是不是一条条对比
// 从数据库里面一条条查询
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		if v.highLoad.Load() {
			// 挂起
		}
		dbctx, cancel := context.WithTimeout(ctx, time.Second)
		var src T
		err := v.base.WithContext(dbctx).
			Where("utime > ?", v.utime).
			Offset(offset).
			Order("utime").First(&src).Error
		cancel()
		switch {
		case errors.Is(err, nil):
			var dst T
			err1 := v.target.WithContext(ctx).
				Where("id = ?", src.ID()).
				First(&dst).Error
			// 此时怎么办
			switch {
			case errors.Is(err1, nil):
				// 找到了，你要开始比较
				//if reflect.DeepEqual(src, dst) {
				//
				//}
				if !src.CompareTo(dst) {
					// 不相等
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				}
			case errors.Is(err1, gorm.ErrRecordNotFound):
				// target 里面少了数据
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				// 要不要汇报数据不一致
				// 1. 大概率一致，记录日志
				v.l.Error("查询 target 数据失败", logger.Error(err1))
				// 2. 出于保险起见，我应该报数据不一致，我去修一下。
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 比完了，没数据了，全量校验结束
			// 同时支持全量校验和增量校验，这里就不能直接返回了
			// 你要考虑，有些情况下，用户希望退出，有些情况下，用户希望继续
			// 当用户希望继续的时候，你要 sleep 一下，然后继续
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		default:
			v.l.Error("校验数据，查询 base 出错", logger.Error(err))
			// offset 位置
		}
		offset++
	}
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	// 先找 target，再找 base
	offset := 0
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)

		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime >?", v.utime).
			Offset(offset).Limit(v.batchSize).
			Order("utime").Find(&dstTs).Error
		if len(dstTs) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		cancel()
		switch {
		// 正常来说，返回结果是多个的情况下，不会进来这个分支
		// gorm 在 Find 的时候，不会返回 gorm.ErrRecordNotFound
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 没数据了，直接返回
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		case errors.Is(err, nil):
			ids := slice.Map(dstTs, func(idx int, src T) int64 {
				return src.ID()
			})
			var srcTs []T
			err = v.base.WithContext(ctx).Where("id IN ?", ids).Find(&srcTs).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				v.notifyBaseMissing(ctx, ids)
			case errors.Is(err, nil):
				srcIds := slice.Map(srcTs, func(idx int, src T) int64 {
					return src.ID()
				})
				// 计算差集 也就是 src 里面没有的
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)

			default:
			}
		default:
			// 记录日志
			v.l.Error("校验数据，查询 target 出错", logger.Error(err))
		}
		offset += len(dstTs)
		if len(dstTs) < v.batchSize {
			// 没数据了
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	cancel()
	if err != nil {
		// 这里怎么办 上报出错，重试？重试也会失败？记日志
		// 我直接忽略，下一轮也会找出来
		v.l.Error("发送数据不一致的消息失败", logger.Error(err))
	}
}
