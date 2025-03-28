package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service

	idx uint64
}

func NewFailoverSMSService(svcs []sms.Service) *FailoverSMSService {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, biz, args, numbers...)
		if err == nil {
			return nil
		}
		// 正常这边，输出日志
		// 要做好监控
		log.Println(err)
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}

func (f *FailoverSMSService) SendV1(ctx context.Context, tplID string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplID, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return err
		default:
			log.Println(err)
		}
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}
