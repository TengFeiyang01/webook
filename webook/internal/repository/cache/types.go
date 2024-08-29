package cache

import (
	"context"
	"webook/webook/internal/domain"
)

// CodeCache 提取为一个接口，所有实现了 Set 和 Verify 的都可以接入
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

// UserCache 和用户操作相关的缓存接口
type UserCache interface {
	Set(ctx context.Context, u domain.User) error
	Get(ctx context.Context, id int64) (domain.User, error)
}
