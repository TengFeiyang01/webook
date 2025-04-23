package startup

import (
	"github.com/IBM/sarama"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}
