package dao

import (
	"gorm.io/gorm"
	dao2 "webook/webook/interactive/repository/dao"
	"webook/webook/internal/repository/dao/article"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&article.Article{},
		&article.PublishedArticleV1{},
		&AsyncSms{},
		&dao2.Interactive{},
		&dao2.UserLikeBiz{},
		&dao2.UserCollectionBiz{},
	)
}
