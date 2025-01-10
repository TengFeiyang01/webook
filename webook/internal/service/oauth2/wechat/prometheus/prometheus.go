package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/service/oauth2/wechat"
)

type Decorator struct {
	svc wechat.Service
	sum prometheus.Summary
}

func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		svc: svc,
		sum: sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.svc.VerifyCode(ctx, code)
}
