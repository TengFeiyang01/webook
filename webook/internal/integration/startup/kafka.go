package startup

import (
	"github.com/IBM/sarama"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9092"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}
