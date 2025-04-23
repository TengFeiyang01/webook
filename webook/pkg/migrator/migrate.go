package migrator

type Entity interface {
	// ID 要求返回 ID
	ID() int64
	// CompareTo dst 必然也是 Entity，正常来说类型是一样的
	CompareTo(dst Entity) bool
}
