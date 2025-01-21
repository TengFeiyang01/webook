package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/webook/internal/service/sms"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	return &PrometheusDecorator{
		svc: svc,
		vector: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: "ytf",
			Subsystem: "webook",
			Name:      "sms_resp_time",
			Help:      "统计 SMS 服务的性能数据",
		}, []string{"biz"}),
	}
}

func (p *PrometheusDecorator) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
