package service

import (
	"context"
	"fmt"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms"
	"math/rand"
)

//go:generate mockgen -source=./code.go -package=svcmocks -destination=./mocks/code.mock.go CodeService
type CodeService interface {
	Send(ctx context.Context,
		// 区别使用业务
		biz string,
		// 这个码, 谁来管, 谁来生成？
		phone string) error
	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

const codeTplID = "123125125"

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context,
	// 区别使用业务
	biz string,
	// 这个码, 谁来管, 谁来生成？
	phone string) error {
	// 生成一个验证码，发送出去
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplID, []string{code}, phone)
	if err != nil {
		// Redis 有这个验证码, 但是没发成功, 用户根本收不到
		// 不能删掉这个验证码
		// 这个 err 可能是超时的 err, 你都不知道发出去了没
		// 要重试的话, 初始化的时候，传入一个自己会重试的
	}
	return err
}

func (svc *codeService) Verify(ctx context.Context, biz string,
	phone string, inputCode string) (bool, error) {
	// phone_code:login:$biz:123456 Redis 存储
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *codeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
