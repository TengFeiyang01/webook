package ratelimit

import (
	"context"
	"fmt"
	"webook/webook/internal/service/sms"
	"webook/webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (r RatelimitSMSService) Send(ctx context.Context, tplID string, args []string, number ...string) error {
	limited, err := r.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误 限流
		// 可以限流：保守策略, 你的下游很坑的时候
		// 可以不限：你的下游很强，业务可用性要求很高, 尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题, %w", err)
	}
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplID, args, number...)
}
