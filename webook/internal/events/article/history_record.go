package article

import (
	"github.com/IBM/sarama"
	"golang.org/x/net/context"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	"webook/webook/pkg/logger"
	"webook/webook/pkg/saramax"
)

type HistoryRecordConsumer struct {
	client sarama.Client
	repo   repository.HistoryRecordRepository
	l      logger.LoggerV1
}

func NewHistoryRecordConsumer(client sarama.Client, repo repository.HistoryRecordRepository, l logger.LoggerV1) *HistoryRecordConsumer {
	return &HistoryRecordConsumer{client: client, repo: repo, l: l}
}

func (r *HistoryRecordConsumer) Start() error {
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
func (r *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage,
	event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})
}
