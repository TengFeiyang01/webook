package article

import (
	"github.com/IBM/sarama"
	"github.com/TengFeiyang01/webook/webook/article/events"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/saramax"
	"golang.org/x/net/context"
	"time"
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
			saramax.NewHandler[events.ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

// Consume 这个不是幂等的
func (r *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage,
	event events.ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "art",
		Uid:   event.Uid,
	})
}
