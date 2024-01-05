package time

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// 用 ticker
	tm := time.NewTicker(time.Second)
	// 这一句不要忘了
	// 避免潜在的 goroutine 泄露的问题
	defer tm.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Log("超时了，或者被取消了")
			// break 不会退出循环
			goto end
			//break end
		case now := <-tm.C:
			t.Log(now.Unix())
		}
	}
end:
	t.Log("退出循环")
}

func TestTimer(t *testing.T) {
	// 精确到 12:00 怎么用 timer
	tm := time.NewTimer(time.Second)
	defer tm.Stop()
	go func() {
		for now := range tm.C {
			t.Log(now.Unix())
		}
	}()

	time.Sleep(time.Second * 10)
}
