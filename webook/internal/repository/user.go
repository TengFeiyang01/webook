package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/cache/user"
	"webook/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

// CachedUserRepository 接收的都是接口
type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}
func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	// 先从cache找
	// 再从dao里面找
	// 找到了写回cache
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	if errors.Is(err, user.ErrKeyNotExist) {
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

	u = r.entityToDomain(ue)

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 打日志, 做监控
		}
	}()

	return u, nil
}

func (r *CachedUserRepository) UpdateByID(ctx context.Context, id int64, u domain.User) error {
	err := r.dao.UpdateByID(ctx, id, r.domainToEntity(u))
	return err
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		ID: u.ID,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		ID:       u.ID,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
