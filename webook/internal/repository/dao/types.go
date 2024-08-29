package dao

import "context"

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
	FindByID(ctx context.Context, id int64) (User, error)
	UpdateByID(ctx context.Context, id int64, u User) error
}
