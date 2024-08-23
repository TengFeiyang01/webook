package repository

import (
	"context"
	"errors"
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

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:       u.ID,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	// 先从cache找
	// 再从dao里面找
	// 找到了写回cache
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	if errors.Is(err, cache.ErrKeyNotExist) {
		// 去数据库加载
	}

	// 这里怎么办? err = io.EOF

	// 加载 —— 做好兜底, 万一 Redis 真的崩了, 你要保护住你的数据库
	// 我数据库限流

	// 不加载 —— 用户体验差一些

	ue, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		ID:       ue.ID,
		Email:    ue.Email,
		Password: ue.Password,
	}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 打日志, 做监控
		}
	}()

	return u, nil
}
