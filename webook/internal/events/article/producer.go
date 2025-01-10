package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"golang.org/x/net/context"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, event ReadEvent) error
	ProduceReadEventV1(ctx context.Context, event ReadEventV1) error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func (k *KafkaProducer) ProduceReadEventV1(ctx context.Context, event ReadEventV1) error {
	//TODO implement me
	panic("implement me")
}

// ProduceReadEvent 如果你有很复杂的重试逻辑, 就用装饰器
func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, event ReadEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func NewKafkaProducer(pc sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: pc,
	}
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type ReadEventV1 struct {
	Uid []int64
	Aid []int64
}
