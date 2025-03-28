package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	// 你的服务商
	svcs []sms.Service
	idx  int32
	// 连续超时的个数
	cnt int32

	// 阈值
	// 连续超时超过或者数字，就要切换
	threshold int32
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context,
	biz string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt >= t.threshold {
		// 这里要切换新的下标
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 如果切换成功
			atomic.StoreInt32(&t.cnt, 0)

		}
		// else 出现并发了，别人换成功了
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, biz, args, numbers...)
	switch {
	case err == nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&t.cnt, 1)
		return err
	default:
		// 不知道什么错误
		// 你可以考虑，换下一个
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		log.Println("failover failover service:", err)
		return err
	}
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}
