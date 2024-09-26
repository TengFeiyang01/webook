package dao

import (
	"gorm.io/gorm"
	"webook/webook/internal/repository/dao/article"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&article.Article{},
		&article.PublishArticle{},
		&AsyncSms{})
}
