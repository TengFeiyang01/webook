package job

import (
	"github.com/google/uuid"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"sync"
	"sync/atomic"
	"time"
	"webook/webook/internal/service"
	"webook/webook/pkg/logger"
)

type LoadBalanceJob struct {
	svc     service.RankingService
	l       logger.LoggerV1
	timeout time.Duration
	client  *rlock.Client
	key     string

	localLock *sync.Mutex
	lock      *rlock.Lock

	// 随机生成一个
	load *atomic.Int32

	nodeID         string
	redisClient    redis.Cmdable
	rankingLoadKey string
	closeSignal    chan struct{}
	loadTicker     *time.Ticker
}

func NewLoadBalanceJob(svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration,
	redisClient redis.Cmdable,
	loadInterval time.Duration) *LoadBalanceJob {
	res := &LoadBalanceJob{
		svc:       svc,
		l:         l,
		timeout:   timeout,
		client:    client,
		localLock: &sync.Mutex{},

		nodeID:         uuid.New().String(),
		redisClient:    redisClient,
		rankingLoadKey: "ranking_job_nodes_load",
		load:           &atomic.Int32{},
		closeSignal:    make(chan struct{}),
		loadTicker:     time.NewTicker(loadInterval),
	}
	res.loadCycle()
	return res
}

func (j *LoadBalanceJob) Name() string {
	return "ranking"
}

func (j *LoadBalanceJob) Run() error {
	j.localLock.Lock()
	lock := j.lock
	j.localLock.Unlock()
	if lock == nil {
		// 我能不能在这里，看一眼我是不是负载最低的，如果是，我就尝试获取分布式锁
		// 如果我的负载低于 70% 的节点

		// 抢夺分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := j.client.Lock(ctx, j.key, j.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
			}, time.Second)
		if err != nil {
			j.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		j.l.Debug(j.nodeID + "获得了分布式锁")
		j.lock = lock
		go func() {
			// 并不是非得一半就续约
			// 如果是自己手写的自动续约，那么可以在续约的时候检查一下负载
			err1 := lock.AutoRefresh(j.timeout/2, j.timeout)
			if err1 != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				j.localLock.Lock()
				j.lock = nil
				j.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()
	return j.svc.TopN(ctx)
}

func (j *LoadBalanceJob) loadCycle() {
	go func() {
		for range j.loadTicker.C {
			j.reportLoad()
			j.releaseLockIfNeed()
		}
	}()
}

func (j *LoadBalanceJob) reportLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	load := j.load.Load()
	j.l.Debug(j.nodeID+"上报负载：", logger.Int32("load", load))
	j.redisClient.ZAdd(ctx, j.rankingLoadKey, redis.Z{Score: float64(load), Member: j.nodeID})
	cancel()
	return
}

func (j *LoadBalanceJob) releaseLockIfNeed() {
	// 检测自己是不是负载最低，如果不是，那么就直接释放分布式锁。
	j.localLock.Lock()
	lock := j.lock
	j.localLock.Unlock()
	if lock != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		// 最低负载的
		// 这里是有一个优化的
		// 你可以说，
		// 1. 如果我的负载高于平均值，就释放分布式锁
		// 2. 如果我的负载高于中位数，就释放分布式锁
		// 3. 如果我的负载高于 70%，就释放分布式锁
		res, err := j.redisClient.ZPopMin(ctx, j.rankingLoadKey).Result()
		if err != nil {
			return
		}
		head := res[0]
		if head.Member.(string) != j.nodeID {
			// 不是自己, 释放锁
			j.l.Debug(j.nodeID+"不是负载最低的节点, 释放分布式锁",
				logger.Field{Key: "head", Value: head})
			j.localLock.Lock()
			j.lock = nil
			j.localLock.Unlock()
			_ = lock.Unlock(ctx)
		}
	}
}

func (j *LoadBalanceJob) Close() error {
	j.localLock.Lock()
	lock := j.lock
	j.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var err *multierror.Error
	if lock != nil {
		err = multierror.Append(err, lock.Unlock(ctx))
	}
	if j.loadTicker != nil {
		j.loadTicker.Stop()
	}
	// 删除自己的负载
	err = multierror.Append(err, j.redisClient.ZRem(ctx, j.rankingLoadKey, redis.Z{Member: j.nodeID}).Err())
	return err
}
