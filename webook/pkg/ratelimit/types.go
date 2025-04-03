package ratelimit

import "context"

type Limiter interface {
	// Limit 有没有触发限流 key 就是限流对象
	// bool 代表是否限流
	Limit(ctx context.Context, key string) (bool, error)
}
