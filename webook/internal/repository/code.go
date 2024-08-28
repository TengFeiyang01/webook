package repository

import (
	"context"
	"webook/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository struct {
	cache *cache.CodeRedisCache
}

func NewCodeRepository(c *cache.CodeRedisCache) *CodeRepository {
	return &CodeRepository{
		cache: c,
	}
}

// 基于内存的实现
//type CodeRepository struct {
//	cache *cache.CodeMemoryCache
//}
//
//func NewCodeRepository(c *cache.CodeMemoryCache) *CodeRepository {
//	return &CodeRepository{
//		cache: c,
//	}
//}

func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, code)
}
