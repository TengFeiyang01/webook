package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]interface{}{
		"utime":     time.Now().UnixMilli(),
		"next_time": next.UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]interface{}{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]interface{}{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) Release(ctx context.Context, id int64) error {
	// 要不要检测 status 或者 version
	// 要的
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]interface{}{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 高并发情况, 大部分都是陪太子读书
	for {
		now := time.Now().UnixMilli()
		var j Job
		// 分布式任务调度系统
		// 1. 一次拉一批, 我一次性取出 100 条, 然后, 我随机从某一条开始, 向后开始抢占
		// 2. 我搞个随机偏移量, 0-100 生成一个随机偏移量。兜底：第一轮没查到, 偏移量回到 0
		// 3. 我搞一个 id 取余分配, status = ? AND next_time <= AND id % 10 = ? 兜底：不加取余 取最老的 next_time
		ddl := now - (time.Minute * 3).Milliseconds()
		err := g.db.WithContext(ctx).Model(&Job{}).
			Where("status = ? AND next_time < ? OR (status = ? AND utime < ?)",
				jobStatusWaiting, now, jobStatusRunning, ddl).First(&j).Error
		// 你找到了, 可以被抢占的, 我要开始抢占了
		if err != nil {
			// 没有任务, 从这里返回
			return Job{}, err
		}
		// 乐观锁, CAS 操作, compare AND Swap
		// 面试刷亮点: 使用乐观锁取代 FOR UPDATE
		// 曾经用了 FOR UPDATE => 性能差、还会有死锁 => 我又化成了乐观锁
		res := g.db.WithContext(ctx).Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).Updates(map[string]interface{}{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			// 抢占失败, 你只能说, 我要继续下一轮
			continue
		}
		return j, nil
	}
}

type Job struct {
	Id       int64 `gorm:"primary_key,AUTO_INCREMENT"`
	Cfg      string
	Name     string `gorm:"unique"`
	Executor string

	// 哪些任务可以抢占, 哪些任务已经被人占着, 哪些任务永远不会执行
	Status int
	// Cron 表达式
	Expression string

	Version int

	// 定时任务, 我怎么知道, 已经到时间了呢？
	// NextTime 下一次被调度的时间
	// next_time <= now && status = 0
	// next_time 和 status 联合索引
	NextTime int64 `gorm:"index"`

	Ctime int64
	Utime int64
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
	jobStatusPaused
)
