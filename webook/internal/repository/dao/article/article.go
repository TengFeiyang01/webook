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
	Upsert(ctx context.Context, art PublishedArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where(" id = ? AND author_id = ?", id, author).
			Updates(map[string]interface{}{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么 ID 是错误的，要么作者不对
			// 后者情况需要小心
			// 用 Prometheus 打点，只要频繁出现，你就告警，然后手工介入排查
			return fmt.Errorf("可能有人在搞你，误操作非自己的文章, uid: %d, author: %d", id, author)
		}
		return tx.Model(&Article{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status": status,
				"utime":  now,
			}).Error
	})
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
		return txDAO.Upsert(ctx, PublishedArticle{Article: art})
	})
	return id, err
}

// Upsert Update or INSERT
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	// 这个是插入
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
			"status":  art.Status,
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
			"status":  art.Status,
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
	Status   uint8
	// - 按 创建时间/更新时间 倒序排序
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC
	// - 在 author_id 和 ctime 上创建联合索引
	Ctime int64 `gorm:"index=aid_ctime"`
	Utime int64
}
