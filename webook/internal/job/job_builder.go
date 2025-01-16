package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"strconv"
	"time"
	"webook/webook/pkg/logger"
)

type CronJobBuilder struct {
	p      *prometheus.SummaryVec
	l      logger.LoggerV1
	tracer trace.Tracer
}

func NewCronJobBuilder(l logger.LoggerV1) *CronJobBuilder {
	p := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      "cron_job",
		Namespace: "ytf",
		Subsystem: "webook",
		Help:      "统计 定时任务 的执行情况",
	}, []string{"name"})
	prometheus.MustRegister(p)

	return &CronJobBuilder{
		l:      l,
		p:      p,
		tracer: otel.GetTracerProvider().Tracer("webook/internal/job"),
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		_, span := b.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		var success bool
		defer func() {
			b.l.Info("任务结束", logger.String("name", name), logger.String("job", job.Name()))
			duration := time.Since(start).Milliseconds()
			b.p.WithLabelValues(name,
				strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run()
		success = err == nil
		if err != nil {
			span.RecordError(err)
			b.l.Error("运行任务失败", logger.Error(err), logger.String("job", job.Name()))
		}

		return nil
	})
}

type cronJobFuncAdapter func() error

func (c cronJobFuncAdapter) Run() {
	_ = c()
}
