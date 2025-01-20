package main

import (
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestCronExpression(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	// 这个任务的标识符
	// @every 是便利语法
	id, err := expr.AddJob("@every 1s", moyJob{})
	expr.AddFunc("@every 3s", func() {
		t.Log("long work start!")
		time.Sleep(6 * time.Second)
		t.Log("long work end!")
	})
	if err != nil {
		return
	}
	require.NoError(t, err)
	t.Log(id)
	expr.Start()
	time.Sleep(12 * time.Second)
	// 发出停止信号, expr 不会调度新的任务, 但是也不会中断已经调度了的任务
	ctx := expr.Stop()
	<-ctx.Done()
}

type moyJob struct {
}

func (m moyJob) Run() {
	log.Println("run")
}
