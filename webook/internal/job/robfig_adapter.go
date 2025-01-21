package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/webook/pkg/logger"
)

type RankingJobAdapter struct {
	j Job
	l logger.LoggerV1
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.LoggerV1) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "cron_job",
		ConstLabels: prometheus.Labels{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{j: j, l: l, p: p}
}

func (r RankingJobAdapter) Run() {
	err := r.j.Run()
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		r.p.Observe(float64(duration))
	}()
	if err != nil {
		r.l.Error("运行任务失败", logger.Error(err), logger.String("job", r.j.Name()))
	}
}
