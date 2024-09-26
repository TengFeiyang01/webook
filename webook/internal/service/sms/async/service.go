package async

import (
	"context"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	sms "webook/webook/internal/service/sms"
	"webook/webook/pkg/logger"
)

type Service struct {
	svc sms.Service
	// 转异步，存储发短信请求的 repository
	repo repository.AsyncSmsRepository
	l    logger.LoggerV1

	responseTimes []time.Duration
	threshold     time.Duration // 阈值，比如500ms
}

func NewService(svc sms.Service,
	repo repository.AsyncSmsRepository,
	l logger.LoggerV1) *Service {
	res := &Service{
		svc:           svc,
		repo:          repo,
		l:             l,
		responseTimes: make([]time.Duration, 0),
		threshold:     500 * time.Millisecond, // 设置阈值为500ms
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 异步发送消息
// 这里我们没有设计退出机制，是因为没啥必要
// 因为程序停止的时候，它自然就停止了
// 原理：这是最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息，确保在非常多个实例
	// 比如 k8s 部署了三个 pod，一个请求，只有一个实例能拿到
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		// 执行发送
		// 这个也可以做成配置的
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			// 啥也不需要干
			s.l.Error("执行异步发送短信失败",
				logger.Error(err),
				logger.Int64("id", as.Id))
		}
		res := err == nil
		// 通知 repository 我这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Int64("id", as.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		// 睡一秒。这个你可以自己决定
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 你可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败",
			logger.Error(err))
		time.Sleep(time.Second)
	}
}

// Send 在发送短信时记录响应时间
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	err := s.svc.Send(ctx, tplId, args, numbers...)
	responseTime := time.Since(start)

	// 记录响应时间
	s.recordResponseTime(responseTime)

	if s.needAsync() {
		err := s.repo.Add(ctx, domain.AsyncSms{
			TplId:    tplId,
			Args:     args,
			Numbers:  numbers,
			RetryMax: 3,
		})
		return err
	}

	return err
}

// 提前引导你们，开始思考系统容错问题
// 你们面试装逼，赢得竞争优势就靠这一类的东西
func (s *Service) needAsync() bool {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	// 记录的响应时间少于3次，则暂不判断为异步
	if len(s.responseTimes) < 3 {
		return false
	}

	// 遍历最近几次响应时间，检查是否都超过阈值
	for _, t := range s.responseTimes {
		if t < s.threshold {
			// 如果有一次响应时间小于阈值，则不需要异步
			return false
		}
	}

	// 如果所有响应时间都超过阈值，则需要转为异步
	return true
}

// recordResponseTime 记录最近的响应时间
func (s *Service) recordResponseTime(duration time.Duration) {
	// 将最新的响应时间加入滑动窗口
	s.responseTimes = append(s.responseTimes, duration)

	// 如果滑动窗口中的记录数超过3次，则移除最旧的一次
	if len(s.responseTimes) > 3 {
		s.responseTimes = s.responseTimes[1:]
	}
}
