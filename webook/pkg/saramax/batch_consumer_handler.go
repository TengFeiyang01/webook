package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"golang.org/x/net/context"
	"time"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

type BatchHandler[T any] struct {
	l  logger.LoggerV1
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 用 option 模式设置
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func([]*sarama.ConsumerMessage, []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:             l,
		fn:            fn,
		batchDuration: time.Second,
		batchSize:     10,
	}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 批量消费
	msgs := claim.Messages()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		batch := make([]*sarama.ConsumerMessage, 0, b.batchSize)
		ts := make([]T, 0, b.batchSize)
		var done = false
		for i := 0; i < b.batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时了
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("failed to unmarshal message",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset))
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		if len(batch) == 0 {
			continue
		}
		// 调用
		err := b.fn(batch, ts)
		if err != nil {
			b.l.Error("Failed to call business batch interface",
				logger.Error(err))
			// 整个批次都记录（没必要）
			// 还有继续往前消费
		}
		// 凑够了一批，然后你就处理
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}
