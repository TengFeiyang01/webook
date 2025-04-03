package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

type HandlerV1[T any] struct {
	l      logger.LoggerV1
	fn     func(msg *sarama.ConsumerMessage, t T) error
	vector *prometheus.SummaryVec
}

func NewHandlerV[T any](consumer string, l logger.LoggerV1, fn func(*sarama.ConsumerMessage, T) error) *HandlerV1[T] {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      "saramax",
		Subsystem: "consumer_handler",
		Namespace: consumer,
	}, []string{"topic", "error"})
	return &HandlerV1[T]{
		l:      l,
		fn:     fn,
		vector: vector,
	}
}

func (h HandlerV1[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h HandlerV1[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h HandlerV1[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

func (h *HandlerV1[T]) consumeClaim(msg *sarama.ConsumerMessage) error {
	start := time.Now()
	var err error
	defer func() {
		errInfo := strconv.FormatBool(err != nil)
		duration := time.Since(start).Milliseconds()
		h.vector.WithLabelValues(msg.Topic, errInfo).Observe(float64(duration))
	}()
	var t T
	err = json.Unmarshal(msg.Value, &t)
	if err != nil {
		h.l.Error("failed to unmarshal message",
			logger.Error(err),
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset))
		return err
	}
	err = h.fn(msg, t)
	if err != nil {
		h.l.Error("failed to consume message",
			logger.Error(err),
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset))
		return err
	}
	return nil
}
