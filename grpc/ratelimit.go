package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

type CounterLimiter struct {
	cnt       *atomic.Int32
	threshold int32
}

// NewCounterLimiter 计数器算法
func (l *CounterLimiter) NewCounterLimiter() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		cnt := l.cnt.Load()
		defer func() {
			l.cnt.Add(1)
		}()
		if cnt > l.threshold {
			return nil, status.Errorf(codes.ResourceExhausted, "grpc rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

type FixedWindowLimiter struct {
	window    time.Duration
	lastStart time.Time
	cnt       int
	threshold int

	lock sync.Mutex
}

func (l *FixedWindowLimiter) NewFixedWindowLimiter() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.lock.Lock()
		now := time.Now()
		if now.Sub(l.lastStart) > l.window {
			l.lastStart = now
			l.cnt = 0
		}
		l.cnt++
		if l.cnt <= l.threshold {
			l.lock.Unlock()
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.ResourceExhausted, "grpc rate limit exceeded")
	}
}

type SlidingWindowLimiter struct {
	window    time.Duration
	queue     queue.PriorityQueue[time.Time]
	lock      sync.Mutex
	threshold int
}

func (l *SlidingWindowLimiter) NewSlidingWindowLimiter() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.lock.Lock()
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(time.Now())
			l.lock.Unlock()
			return handler(ctx, req)
		}
		for {
			first, _ := l.queue.Peek()
			if first.Before(time.Now().Add(-l.window)) {
				_, _ = l.queue.Dequeue()
			} else {
				break
			}
		}
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(time.Now())
			l.lock.Unlock()
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.ResourceExhausted, "grpc rate limit exceeded")
	}
}

type TokenBucketLimiter struct {
	//ch       *time.Ticker
	buckets  chan struct{}
	interval time.Duration

	closeCh chan struct{}
}

// NewTokenBucketLimiter 把 capacity 设置成0， 就变成了漏桶算法
func NewTokenBucketLimiter(interval time.Duration, capacity int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets:  make(chan struct{}, capacity),
		interval: interval,
	}
}

func (l *TokenBucketLimiter) NewServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(l.interval)
	go func() {
		for {
			select {
			case <-l.closeCh:
				return
			case <-ticker.C:
				l.buckets <- struct{}{}
			}
		}
	}()
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-l.buckets:
			return handler(ctx, req)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (l *TokenBucketLimiter) Close() {
	close(l.closeCh)
}
