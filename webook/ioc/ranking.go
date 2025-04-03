package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
	"github.com/TengFeiyang01/webook/webook/internal/job"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1, rlockClient *rlock.Client) *job.RankingJob {
	// todo: 暴露出来 job.Close()
	return job.NewRankingJob(svc, time.Second*30, rlockClient, l)
}

func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
