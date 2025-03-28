package events

import (
	"github.com/IBM/sarama"
	"golang.org/x/net/context"
	"time"
	"github.com/TengFeiyang01/webook/webook/interactive/repository"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/saramax"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l}
}

func (r *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

// Consume 这个不是幂等的
func (r *InteractiveReadEventBatchConsumer) Consume(message []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "art")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := r.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		r.l.Error("failed to batch incr read_cnt",
			logger.Field{Key: "ids", Value: ids},
			logger.Field{Key: "bizs", Value: bizs},
			logger.Error(err))
	}
	return nil
}
