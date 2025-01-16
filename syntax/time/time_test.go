package time

import (
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// 间隔
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
o:
	for {
		select {
		case <-ticker.C:
			t.Log(time.Now())
		case <-ctx.Done():
			t.Log("超时了")
			break o
			//goto end
		}
	}
	//end:
	t.Log("退出循环")
}

func TestTimer(t *testing.T) {
	// 定时器
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	go func() {
		for now := range timer.C {
			t.Log(now.Unix())
		}
	}()
	time.Sleep(time.Second)
}
