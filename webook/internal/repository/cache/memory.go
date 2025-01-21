package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

// CodeCache 提取为一个接口，所有实现了 Set 和 Verify 的都可以接入
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

var ErrKeyNotExist = errors.New("key not exist")

type LocalCodeCache struct {
	lock       sync.Mutex
	cache      *lru.Cache
	rwLock     sync.RWMutex
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

type codeItem struct {
	code   string
	cnt    int
	expire time.Time
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	now := time.Now()
	val, ok := l.cache.Get(key)
	if !ok {
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	item, ok := val.(codeItem)
	if !ok {
		return errors.New("系统错误")
	}
	if item.expire.Sub(now) > time.Minute*9 {
		return ErrCodeSendTooMany
	}
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		return false, ErrKeyNotExist
	}
	item, ok := val.(codeItem)
	if !ok {
		return false, errors.New("系统错误")
	}
	if item.cnt <= 0 {
		return false, ErrCodeVerifyTooManyTimes
	}
	item.cnt--
	return item.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (l *LocalCodeCache) cntKey(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s:cnt", biz, phone)
}
