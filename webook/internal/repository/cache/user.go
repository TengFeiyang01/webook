package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

// UserCache 和用户操作相关的缓存接口
type UserCache interface {
	Set(ctx context.Context, u domain.User) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Del(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	// 传单机 Redis 可以
	// 传单机 cluster 的 Redis 也可以
	cmd        redis.Cmdable
	expiration time.Duration
}

// A 用到了 B, B一定是接口
// A 用到了 B, B一定是 A 的字段
// A 用到了 B, A 绝对不初始化 B, 而是从外面传入

func NewRedisUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        client,
		expiration: time.Minute * 15,
	}
}

// Get 只要 err 为 nil, 就认为有缓存有数据
// 如果没有数据, 返回一个特定的 error
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// 数据不存在
	result, err := cache.cmd.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(result, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.ID)
	return cache.cmd.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) Del(ctx context.Context, id int64) error {
	return cache.cmd.Del(ctx, cache.key(id)).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
