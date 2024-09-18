package retryable

import (
	"context"
	"webook/webook/internal/service/sms"
)

// Service 小心并发问题
type Service struct {
	svc      sms.Service
	retryCnt int
}

func (s Service) Send(ctx context.Context, tplID string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tplID, args, numbers...)
	for err != nil && s.retryCnt < 10 {
		err = s.svc.Send(ctx, tplID, args, numbers...)
		s.retryCnt++
	}
	return err
}
