package repository

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: dao}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:       u.ID,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}

func (r *UserRepository) FindByID(id int64) {
	// 先从cache找
	// 再从dao里面找
	// 找到了写回cache
}
