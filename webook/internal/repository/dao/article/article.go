package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	// 在事务内部、这里采用了闭包形态
	// GORM 帮助我们管理事务的生命周期
	var (
		id = art.Id
	)
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 操作线上库了
		return txDAO.Upsert(ctx, PublishArticle{Article: art})
	})
	return id, err
}

// Upsert Update or INSERT
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishArticle) error {
	// 这个是插入
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		}),
	}).Create(&art).Error
	// INSERT xxx on DUPLICATE KEY UPDATE xxx
	return err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	art.Utime = time.Now().UnixMilli()
	// 依赖 gorm 忽略零值, 会用主键进行更新
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		})
	// 你要不要检查真的更新了
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update article failed, the author maybe invalid. id:[%d], author_id:[%d]", art.Id, art.AuthorId)
	}
	return res.Error
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
