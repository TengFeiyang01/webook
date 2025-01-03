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
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	UpdateById(ctx context.Context, u domain.User) error
	UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error
}

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	WithDraw(ctx context.Context, art domain.Article) error
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}
