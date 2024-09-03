package code

import (
	"context"
	"errors"
	"fmt"
	memcache "github.com/patrickmn/go-cache"
	"time"
	"webook/webook/internal/repository/cache"
)

const expiration = 10 * time.Minute

type MemoryCodeCache struct {
	cache *memcache.Cache
}

func NewMemoryCodeCache() cache.CodeCache {
	return &MemoryCodeCache{
		cache: memcache.New(expiration, expiration),
	}
}

func (mc *MemoryCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	_ = ctx
	_, t, ok := mc.cache.GetWithExpiration(mc.key(biz, phone))
	if ok && time.Now().After(t) {
		return errors.New("系统错误")
	}
	if !ok || t.Sub(time.Now()) < time.Minute*9 {
		mc.cache.Set(mc.key(biz, phone), code, expiration)
		mc.cache.Set(mc.cntKey(biz, phone), 3, expiration)
		return nil
	}
	return ErrCodeSendTooMany
}

func (mc *MemoryCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	_ = ctx
	cntKey := mc.cntKey(biz, phone)
	cnt, ok := mc.cache.Get(cntKey)
	if !ok || cnt.(int) <= 0 {
		return false, ErrCodeVerifyTooManyTimes
	}
	code, _ := mc.cache.Get(mc.key(biz, phone))
	if inputCode == code {
		return true, nil
	} else {
		_, err := mc.cache.DecrementInt(mc.key(biz, phone), 1)
		return false, err
	}
}

func (mc *MemoryCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (mc *MemoryCodeCache) cntKey(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s:cnt", biz, phone)
}
