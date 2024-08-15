package domain

import "time"

// User 领域对象 是 DDD 中的 entity/聚合根
// BO(Business object)
type User struct {
	Email    string
	Password string
	Ctime    time.Time
}
