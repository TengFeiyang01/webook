package cache

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
	"github.com/TengFeiyang01/webook/webook/article/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, author int64) error

	Set(ctx context.Context, id int64, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	SetPub(ctx context.Context, id int64, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error) {
	bs, err := r.client.Get(ctx, r.firstPageKey(author)).Result()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal([]byte(bs), &arts)
	return arts, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.firstPageKey(author), data, time.Minute*10).Err()
}

func (r *RedisArticleCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("art:first_page:%d", uid)
}

func (r *RedisArticleCache) key(uid int64) string {
	return fmt.Sprintf("art:%d", uid)
}

func (r *RedisArticleCache) DelFirstPage(ctx context.Context, author int64) error {
	return r.client.Del(ctx, r.firstPageKey(author)).Err()
}

func (r *RedisArticleCache) Set(ctx context.Context, id int64, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 过期时间要短，预测效果越不好，就越要短
	return r.client.Set(ctx, r.key(id), data, time.Second*10).Err()
}

func (r *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	bs, err := r.client.Get(ctx, r.key(id)).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal([]byte(bs), &art)
	return art, err
}

func (r *RedisArticleCache) SetPub(ctx context.Context, id int64, article domain.Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(id), data, time.Minute).Err()
}

func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	bs, err := r.client.Get(ctx, r.key(id)).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal([]byte(bs), &art)
	return art, err
}
