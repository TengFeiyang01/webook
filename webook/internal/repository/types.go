package repository

import (
	"context"
	"webook/webook/internal/domain"
)

// CodeRepository Code相关的功能
type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

// UserRepository 所有的 User 相关的功能
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	UpdateById(ctx context.Context, u domain.User) error
}
