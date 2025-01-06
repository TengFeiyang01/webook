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
	Upsert(ctx context.Context, art PublishedArticleV1) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
	GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (Article, error)
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).First(&art, "id = ?", id).Error
	return art, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author = ?", author).
		Offset(offset).
		Limit(limit).
		// 升序老徐. utime ASC
		Order("utime DESC").
		//Order(clause.OrderByColumn{Column: clause.Column{Name: "utime"}, Desc: true}).
		Find(&arts).Error
	return arts, err
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
		return txDAO.Upsert(ctx, PublishedArticleV1{Article: art})
	})
	return id, err
}

// Upsert Update or INSERT
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticleV1) error {
	// 这个是插入
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
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
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
