package context

import (
	"context"
	"testing"
	"time"
)

type Key1 struct{}

func TestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(),
		Key1{}, "value1")
	val := ctx.Value(Key1{})
	t.Log(val)
	ctx = context.WithValue(ctx, "key2", "value2")
	val = ctx.Value(Key1{})
	t.Log(val)
	val = ctx.Value("key2")
	t.Log(val)
	ctx = context.WithValue(ctx, Key1{}, "value1-1")
	val = ctx.Value(Key1{})
	t.Log(val)
}

func TestContext_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// 为什么一定要 cancel 呢？
	// 防止 goroutine 泄露
	defer cancel()

	// 防止有些人使用了 Done，在等待ctx结束信号
	go func() {
		ch := ctx.Done()
		<-ch
	}()

	// 在这里用 ctx

	ctx = context.WithValue(ctx, Key1{}, "value1-1")
	val := ctx.Value(Key1{})
	t.Log(val)

	ctx, cancel = context.WithTimeout(ctx, time.Second)
	cancel()
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second))
	cancel()
}

func TestContextErr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	ctx.Err()
	// 你怎么区别被取消，还是超时了呢？
	if ctx.Err() == context.Canceled {

	} else if ctx.Err() == context.DeadlineExceeded {

	}
}

func TestContextSub(t *testing.T) {
	ctx, cancel0 := context.WithCancel(context.Background())
	subCtx, _ := context.WithCancel(ctx)

	go func() {
		time.Sleep(time.Second)
		cancel0()
	}()

	go func() {
		// 监听 subCtx 结束的信号
		t.Log("等待信号...")
		<-subCtx.Done()
		t.Log("收到信号...")
	}()
	time.Sleep(time.Second * 10)
}

func TestContextSubCancel(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	_, cancel1 := context.WithCancel(ctx)

	go func() {
		time.Sleep(time.Second)
		cancel1()
	}()

	go func() {
		// 监听 ctx 结束的信号
		t.Log("等待信号...")
		<-ctx.Done()
		t.Log("收到信号...")
	}()
	time.Sleep(time.Second * 10)
}

//func MockIO() {
//	select {
//
//	case <-ctx.Done():
//		// 监听超时 或者用户主动取消
//
//	case <-biz.Signal():
//		// 监听你的正常业务
//	}
//}
