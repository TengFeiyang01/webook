package domain

import "time"

// User 领域对象 是 DDD 中的 entity/聚合根
// BO(Business object)
type User struct {
	ID         int64
	Email      string
	Password   string
	Phone      string
	NickName   string
	BirthDay   time.Time
	AboutMe    string
	WechatInfo WechatInfo
	Ctime      time.Time
}
