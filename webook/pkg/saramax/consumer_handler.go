package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"webook/webook/pkg/logger"
)

type Handler[T any] struct {
	l  logger.LoggerV1
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger.LoggerV1, fn func(*sarama.ConsumerMessage, T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}

func (h Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("failed to unmarshal message",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset))
			continue
		}
		err = h.fn(msg, t)
		// 在这里执行重试
		if err != nil {
			h.l.Error("failed to consume message",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset))
			return err
		} else {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
