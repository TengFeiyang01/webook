package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
)

type Executor interface {
	Name() string
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候, 就需要考虑结束执行
	// 具体实现来控制
	// 真正去执行
	Exec(ctx context.Context, j domain.Job) error
}

type HttpExecutor struct {
}

func (h HttpExecutor) Name() string {
	return "http"
}

func (h HttpExecutor) Exec(ctx context.Context, j domain.Job) error {
	type Config struct {
		Endpoint string `json:"endpoint"`
		Method   string `json:"method"`
	}
	var cfg Config
	err := json.Unmarshal([]byte(j.Cfg), &cfg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(cfg.Method, cfg.Endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		return errors.New("执行失败")
	}
	return nil
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
	//fn    func(ctx context.Context, j domain.Job)
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Executor]
	if !ok {
		return fmt.Errorf("未知任务, 你是否注册? %s", j.Executor)
	}
	return fn(ctx, j)
}

type Schedule struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.LoggerV1
	limiter *semaphore.Weighted
}

func NewSchedule(svc service.JobService, l logger.LoggerV1) *Schedule {
	return &Schedule{svc: svc, l: l,
		limiter: semaphore.NewWeighted(200),
		execs:   make(map[string]Executor)}
}

func (s *Schedule) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Schedule) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// main 函数退出
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 你不能 return
			// 你要继续下一轮
			s.l.Error("抢占任务失败", logger.Error(err))
		}

		exec, ok := s.execs[j.Executor]
		if !ok {
			// DEBUG 的时候 最后中断
			s.l.Error("未找到对应的执行器", logger.String("executor", j.Executor))
			continue
		}

		// 接下来就是执行
		// 怎么执行
		go func() {
			defer func() {
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					s.l.Error("释放任务失败", logger.Error(err1),
						logger.Int64("id", j.Id))
				}
			}()
			// 异步执行
			// 这边要考虑超时控制
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				// 考虑在这里重试
				s.l.Error("任务执行失败", logger.Error(err1))
			}
			// 你要不要考虑下一次调度?
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}
		}()
	}
}
