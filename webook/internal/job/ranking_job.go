package job

import (
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.LoggerV1
	lock      *rlock.Lock
	locallock *sync.Mutex
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client, l logger.LoggerV1) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		lock:      &rlock.Lock{},
		locallock: &sync.Mutex{},
	}
}

func (r RankingJob) Name() string {
	return "ranking"
}

// Run 按时间调度，三分钟一次
func (r RankingJob) Run() error {
	r.locallock.Lock()
	defer r.locallock.Unlock()

	if r.lock == nil {
		// 说明你没拿到锁, 你得试着拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这里没拿到, 极大概率是别人已经持有
			return nil
		}
		r.lock = lock
		// 我怎么保证我这里, 一直拿着这个锁
		go func() {
			r.locallock.Lock()
			defer r.locallock.Unlock()
			// 自动续约机制
			err := lock.AutoRefresh(r.timeout/2, time.Second)
			// 这里说明退出了续约机制
			if err != nil {
				r.lock = nil
				return
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r RankingJob) Close() error {
	r.locallock.Lock()
	lock := r.lock
	r.lock = nil
	r.locallock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
