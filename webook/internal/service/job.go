package service

import (
	"context"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
)

//go:generate mockgen -source=./job.go -package=svcmocks -destination=./mocks/job.mock.go JobService
type JobService interface {
	// Preempt 抢占
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	//PreemptV1(ctx context.Context) (domain.Job, func() error, error)
}

type cronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func (p *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		return p.repo.Stop(ctx, j.Id)
	}
	return p.repo.UpdateNextTime(ctx, j.Id, next)
}

func (p *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.repo.Preempt(ctx)

	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for range ticker.C {
			p.refresh(j.Id)
		}
	}()

	// 你抢占之后，你一直抢占吗？
	j.CancelFunc = func() error {
		// 自己在这里释放掉
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return p.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (p *cronJobService) refresh(id int64) {
	// 如何续约？
	// 更新一下更新时间即可
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := p.repo.UpdateUtime(ctx, id)
	if err != nil {
		// 可以考虑重试
		p.l.Error("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}
