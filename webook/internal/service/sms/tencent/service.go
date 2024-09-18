package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"webook/webook/pkg/ratelimit"
)

type Service struct {
	appId     *string
	signature *string
	client    *sms.Client
	limiter   ratelimit.Limiter
}

func NewService(c *sms.Client, appId string,
	signName string, limiter ratelimit.Limiter) *Service {
	return &Service{
		client:    c,
		appId:     ekit.ToPtr[string](appId),
		signature: ekit.ToPtr[string](signName),
		limiter:   limiter,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SetContext(ctx)
	req.SignName = s.signature
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.PhoneNumberSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) == "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
