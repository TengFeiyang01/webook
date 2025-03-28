package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"github.com/TengFeiyang01/webook/webook/interactive/events"
	"github.com/TengFeiyang01/webook/webook/pkg/saramax"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `json:"addrs" yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c1 *events.InteractiveEventConsumer) []saramax.Consumer {
	return []saramax.Consumer{c1}
}
