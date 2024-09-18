package failover

import (
	"context"
	"errors"
	"log"
	"webook/webook/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
}

func NewFailoverSMSService(svcs []sms.Service) *FailoverSMSService {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, tplID string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplID, args, numbers...)
		if err == nil {
			return nil
		}
		// 正常这边，输出日志
		// 要做好监控
		log.Println(err)
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}
