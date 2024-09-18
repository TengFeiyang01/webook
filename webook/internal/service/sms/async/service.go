package async

import (
	"context"
	"webook/webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	//dao repository.SMSAsyncReqRepository
}

func NewSMSService() *SMSService {
	return &SMSService{}
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	// 首先是正常路径
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		// 判定是不是崩溃
		//
	}
	panic("")
}
