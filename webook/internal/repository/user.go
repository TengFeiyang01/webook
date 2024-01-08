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

func NewUserRepository(d *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: d,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
	// 在这里操作缓存
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 找
	// 再从 dao 里面找
	// 找到了再往回写
	u, err := r.dao.FindById(ctx, id)
	return domain.User{
		Email:    u.Email,
		Password: u.Password,
	}, err
}
