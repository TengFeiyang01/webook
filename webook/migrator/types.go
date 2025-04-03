package migrator

type Entity interface {
	ID() int64
	CompareTo(t Entity) bool
}
