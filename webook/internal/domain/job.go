package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	// 通用的任务的抽象, 我们也不知道任务的具体细节, 所以就搞一个 Cfg
	Id         int64
	Name       string
	Executor   string
	Cfg        string
	Cron       string
	CancelFunc func() error
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func (j Job) NextTime() time.Time {
	// 你怎么算
	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
