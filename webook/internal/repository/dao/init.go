package dao

import (
	"github.com/TengFeiyang01/webook/webook/article/repository/dao"
	dao2 "github.com/TengFeiyang01/webook/webook/interactive/repository/dao"
	"gorm.io/gorm"
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
