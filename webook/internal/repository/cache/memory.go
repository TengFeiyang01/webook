package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

const expiration = 10 * time.Minute

type CodeMemoryCache struct {
	cache *cache.Cache
}

func NewCodeMemoryCache() *CodeMemoryCache {
	return &CodeMemoryCache{
		cache: cache.New(expiration, expiration),
	}
}

func (mc *CodeMemoryCache) Set(ctx context.Context, biz, phone, code string) error {
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

func (mc *CodeMemoryCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
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

func (mc *CodeMemoryCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (mc *CodeMemoryCache) cntKey(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s:cnt", biz, phone)
}
