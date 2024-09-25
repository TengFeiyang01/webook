package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id       int64  `gorm:"primary_key;AUTO_INCREMENT"`
	Title    string `gorm:"type=varchar(2024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index=aid_ctime"`
	// - 按 创建时间/更新时间 倒序排序
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC
	// - 在 author_id 和 ctime 上创建联合索引
	Ctime int64 `gorm:"index=aid_ctime"`
	Utime int64
}
