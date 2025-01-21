package article

import (
	"github.com/IBM/sarama"
	"golang.org/x/net/context"
	"time"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
	"webook/webook/pkg/saramax"
)

type InteractiveEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveEventConsumer {
	return &InteractiveEventConsumer{client: client, repo: repo, l: l}
}

func (r *InteractiveEventConsumer) Start() error {
	// 在这里加入一个监控，上报 Prometheus 即可
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

// Consume 这个不是幂等的
func (r *InteractiveEventConsumer) Consume(message *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.IncrReadCnt(ctx, "article", t.Aid)
}
