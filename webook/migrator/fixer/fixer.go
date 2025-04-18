package fixer

import (
	"errors"
	"github.com/TengFeiyang01/webook/webook/migrator"
	"github.com/TengFeiyang01/webook/webook/migrator/events"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB // 以什么为准
	target  *gorm.DB // 校验的是谁的数据
	columns []string
}

func (f *Fixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := f.target.WithContext(ctx).Where("id =?", evt.ID).First(&t).Error
	switch {
	case errors.Is(err, nil):
		// base 里面有
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		// base 里面没有
		return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
	default:
		return err
	}
}

// FixV1 base 原本的数据，到我修复的时候可能以及变化了
func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing,
		events.InconsistentEventTypeNEQ:
		// 这边要插入
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			// base 也删除了这条数据
			return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
		case err == nil:
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Error
		default:
			return err
		}

	case events.InconsistentEventTypeBaseMissing:
		// 这边要删除
		return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

func (f *Fixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing:
		// 这边要插入
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			// base 也删除了这条数据
			return nil
		case err == nil:
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Error
		default:
			return err
		}
	case events.InconsistentEventTypeNEQ:
		// 这边要更新
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			// base 也删除了这条数据，那我们就删除 target 里面的这条数据
			return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
		case err == nil:
			return f.target.WithContext(ctx).Save(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		// 这边要删除
		return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}
