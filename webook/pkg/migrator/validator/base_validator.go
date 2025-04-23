package validator

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator/events"
	"gorm.io/gorm"
	"time"
)

type baseValidator struct {
	base   *gorm.DB
	target *gorm.DB

	// 这边需要告知，是以 SRC 为准，还是以 DST 为准
	// 修复数据需要知道
	direction string

	l        logger.LoggerV1
	producer events.Producer
}

// 上报不一致的数据
func (v *baseValidator) notify(id int64, typ string) {
	// 这里我们要单独控制超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{
		Direction: v.direction,
		ID:        id,
		Type:      typ,
	}

	err := v.producer.ProduceInconsistentEvent(ctx, evt)
	if err != nil {
		v.l.Error("发送消息失败", logger.Error(err),
			logger.Field{Key: "event", Value: evt})
	}
}
