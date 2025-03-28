package repository

import (
	"golang.org/x/net/context"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	// 使用具体实现, 可读性更好, 对测试不友好
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.local.Get(ctx)
	if err == nil {
		return arts, err
	}
	// 回写本地缓存
	arts, err = c.redis.Get(ctx)
	if err == nil {
		_ = c.local.Set(ctx, arts)
	} else {
		return c.local.ForceGet(ctx)
	}
	return arts, err
}

func NewCachedRankingRepository(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository{
		redis: redis,
		local: local,
	}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.local.Set(ctx, arts)
	return c.redis.Set(ctx, arts)
}
