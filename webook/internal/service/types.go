package service

import (
	"context"
	"webook/webook/internal/domain"
)

type CodeService interface {
	Send(ctx context.Context,
		// 区别使用业务
		biz string,
		// 这个码, 谁来管, 谁来生成？
		phone string) error
	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	UpdateByID(ctx context.Context, id int64, u domain.User) error
}
