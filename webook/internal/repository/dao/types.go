package dao

import "context"

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
	FindById(ctx context.Context, id int64) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindByWechat(ctx context.Context, openID string) (User, error)
}

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
}
