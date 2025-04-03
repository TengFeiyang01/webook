package events

import "context"

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type InconsistentEvent struct {
	ID int64
	// 用什么来修，取值为 src，意味着以源表为主，取值为 dst，意味着以目标表为主
	Direction string
	Type      string
}

const (
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺了这一条
	InconsistentEventTypeTargetMissing = "target_missing"
	InconsistentEventTypeBaseMissing   = "base_missing"
	// InconsistentEventTypeNEQ 不相等
	InconsistentEventTypeNEQ = "neq"
)
