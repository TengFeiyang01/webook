package dao

import (
	"gorm.io/gorm"
	"webook/webook/article/repository/dao"
	dao2 "webook/webook/interactive/repository/dao"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&dao.Article{},
		&dao.PublishedArticleV1{},
		&AsyncSms{},
		&dao2.Interactive{},
		&dao2.UserLikeBiz{},
		&dao2.UserCollectionBiz{},
	)
}
