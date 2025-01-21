package job

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"math/rand"
	"testing"
	"time"
	svcmocks "webook/webook/internal/service/mocks"
	"webook/webook/pkg/logger"
)

func TestRankingJobHomework(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	job1 := newJob("node1", redisClient, ctrl)
	go func() {
		job1.randomLoad()
	}()
	job2 := newJob("node2", redisClient, ctrl)
	go func() {
		job2.randomLoad()
	}()
	job3 := newJob("node3", redisClient, ctrl)
	job3.randomLoad()
}

func newJob(id string, redisClient redis.Cmdable, ctrl *gomock.Controller) *LoadBalanceJob {
	svc := svcmocks.NewMockRankingService(ctrl)
	svc.EXPECT().TopN(gomock.Any()).AnyTimes().Return(nil)
	zl, _ := zap.NewDevelopment()
	l := logger.NewZapLogger(zl)
	job := NewLoadBalanceJob(
		svc,
		l,
		rlock.NewClient(redisClient),
		time.Minute,
		redisClient,
		time.Second*10)
	job.nodeID = id
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			_ = job.Run()
		}
	}()
	return job
}

func (j *LoadBalanceJob) randomLoad() {
	// 模拟负载变化
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		load := rand.Int31n(100)
		j.load.Store(load)
	}
}
