package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	start := time.Now()
	err = consumer.Consume(ctx,
		[]string{"test_topic"}, ConsumerHandler{})
	assert.NoError(t, err)
	t.Log(time.Since(start).String())
}

type ConsumerHandler struct {
}

func (c ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Setup")
	//partitions := session.Claims()["test_topic"]
	//for _, part := range partitions {
	//	session.ResetOffset("test_topic",
	//		part, saramax.OffsetOldest, "")
	//}
	return nil
}

func (c ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Cleanup")
	return nil
}

func (c ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		log.Println("一个批次开始")
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done = false
		var eg errgroup.Group
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时了
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					// 并发处理
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}
		// 凑够了一批，然后你就处理
		// log.Println(batch)

		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}

func (c ConsumerHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		log.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	return nil
}
