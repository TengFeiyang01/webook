package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	"github.com/redis/go-redis/v9"
	"time"
)

// UserRepository 所有的 User 相关的功能
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	UpdateById(ctx context.Context, u domain.User) error
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
	ErrKeyNotExist   = redis.Nil
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
	return r.toDomain(u), nil
}
func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.toEntity(u))
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从cache找
	// 再从dao里面找
	// 找到了写回cache
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	if errors.Is(err, ErrKeyNotExist) {
		// 去数据库加载
	}

	// 这里怎么办? err = io.EOF

	// 加载 —— 做好兜底, 万一 Redis 真的崩了, 你要保护住你的数据库
	// 我数据库限流

	// 不加载 —— 用户体验差一些

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.toDomain(ue)

	//_ = r.cache.Set(ctx, u)
	//if err != nil {
	//	打日志, 做监控
	//}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 打日志, 做监控
		}
	}()

	return u, nil
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	// 先从cache找
	// 再从dao里面找
	// 找到了写回cache
	var u domain.User
	ue, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}

	u = r.toDomain(ue)

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 打日志, 做监控
		}
	}()

	return u, nil
}

func (r *CachedUserRepository) UpdateById(ctx context.Context, u domain.User) error {
	err := r.dao.UpdateById(ctx, r.toEntity(u))
	return err
}

func (r *CachedUserRepository) UpdateNonZeroFields(ctx context.Context,
	user domain.User) error {
	// 更新 DB 之后，删除
	err := r.dao.UpdateById(ctx, r.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟一秒
	time.AfterFunc(time.Second, func() {
		_ = r.cache.Del(ctx, user.ID)
	})
	return r.cache.Del(ctx, user.ID)
}

func (r *CachedUserRepository) toEntity(u domain.User) dao.User {
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
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
		NickName: u.NickName,
		BirthDay: u.BirthDay.UnixMilli(),
		AboutMe:  u.AboutMe,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		ID:       u.ID,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		NickName: u.NickName,
		BirthDay: time.UnixMilli(u.BirthDay),
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenID.String,
			UnionID: u.WechatUnionID.String,
		},
		AboutMe: u.AboutMe,
		Ctime:   time.UnixMilli(u.Ctime),
	}
}
