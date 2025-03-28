package cache

import (
	"encoding/json"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client redis.Cmdable
	key    string
}

func NewRankingRedisCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{client: client, key: "ranking"}
}

func (r RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = ""
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 这个过期时间要稍微长一些, 最好是超过计算热榜的时间 (包含重试在内的时间)
	// 你甚至可以永不过期
	return r.client.Set(ctx, r.key, val, time.Minute*10).Err()
}

func (r RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	dara, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(dara, &arts)
	return arts, nil
}
