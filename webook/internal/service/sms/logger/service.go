package logger

import (
	"context"
	"go.uber.org/zap"
	"webook/webook/internal/service/sms"
)

type Service struct {
	svc sms.Service
}

func (s Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	zap.L().Debug("send sms code", zap.String("biz", biz), zap.Any("args", args))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		zap.L().Debug("failed to send sms code", zap.Error(err))
	}
	return err
}

func NewService() *Service {
	return &Service{}
}
