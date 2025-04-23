package events

type InconsistentEvent struct {
	Type      string
	ID        int64
	Direction string
}

const (
	// InconsistentEventTypeTargetMissing target 中没有数据
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNotEqual 目标表和源表的数据不相等
	InconsistentEventTypeNotEqual = "neq"
	// InconsistentEventTypeBaseMissing base 中没有数据
	InconsistentEventTypeBaseMissing = "base_missing"
)
