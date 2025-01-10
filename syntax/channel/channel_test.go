package channel

import (
	"golang.org/x/net/context"
	"sync"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	// 声明了一个放 int 类型的 channel
	// 没有初始化, 读写 ch 都会崩溃
	//var ch chan int
	// 不带容量的要谨慎
	//ch1 := make(chan int)
	// 带缓冲区的
	ch2 := make(chan int, 2)
	ch2 <- 123
	close(ch2)
	val, ok := <-ch2
	if !ok {
		// ch2 已经被关闭
	}
	println(val, ok)
}

func TestChannelClose(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 123
	ch <- 234
	val, ok := <-ch
	println(val, ok)
	close(ch)
	//ch <- 1 // send on closed channel [recovered]
	val, ok = <-ch
	println(val, ok)
}

type MyStruct struct {
	ctx       context.Context
	ch        chan struct{}
	closeOnce sync.Once
}

// 用户会多次调用,或者多个 goroutine 调用
func (m *MyStruct) Close() error {
	m.closeOnce.Do(func() {
		close(m.ch)
	})
	return nil
}

func (m *MyStruct) CloseV1() error {
	select {
	case <-m.ctx.Done():
		close(m.ch)
	default:
		return nil
	}
	return nil
}

func TestForLoop(t *testing.T) {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			time.Sleep(time.Millisecond * 100)
		}
	}()
	close(ch)
	for val := range ch {
		t.Log(val)
	}
	println("ok")
}

func TestChannelBlock(t *testing.T) {
	ch := make(chan int)
	// 阻塞
	val := <-ch
	// 执行不下去
	t.Log(val)
}

func TestGoroutineCh(t *testing.T) {
	ch := make(chan int)
	go func() {
		// 永久阻塞在这里
		// 内存泄漏
		ch <- 123
	}()
	// 后面没人从 ch 读数据
	//time.Sleep(time.Millisecond * 1000)
}

func TestGoroutineChV1(t *testing.T) {
	ch := make(chan int, 100000)
	go func() {
		// 永久阻塞在这里
		// 内存泄漏
		for i := 0; i < 100000; i++ {
			ch <- i
		}
		ch <- 1
	}()
	// 后面没人从 ch 读数据
}
