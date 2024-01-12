package repository

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(d *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   d,
		cache: c,
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

// FindById
// 缺点：只要缓存返回了 error，就直接取数据库查询。
//
//	回写缓存的时候，忽略掉了错误
func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)

	switch err {
	case nil:
		return u, err
	case cache.ErrKeyNotExist:
		ue, err := r.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}

		u = domain.User{
			Id:       ue.Id,
			Email:    ue.Email,
			Password: ue.Password,
		}
		_ = r.cache.Set(ctx, u)
		return u, nil
	default:
		return domain.User{}, err
	}
}
