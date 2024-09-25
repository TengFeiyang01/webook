package repository

import (
	"context"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/cache/code"
)

// CodeRepository Code相关的功能
type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

var (
	ErrCodeSendTooMany        = code.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = code.ErrCodeVerifyTooManyTimes
)

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

func (repo *CachedCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, code)
}
