package article

import (
	"context"
	"gorm.io/gorm"
	"webook/webook/internal/domain"
	dao "webook/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO

	// V1 操作两个 DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDao

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果你要在 repository 层面上操作事务
	// 那么就只能利用 db 开启事务之后，创建基于事务的 DAO
	// 或者，直接去掉 DAO 这一层、在 repository 的实现中，直接操作 db。
	db *gorm.DB
}

func (c CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

// SyncV2 尝试在 repository 层解决事务问题
func (c CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	// 利用 tx 来构建DAO
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)
	var (
		id  = art.Id
		err error
	)
	artEntity := c.toEntity(art)
	if id > 0 {
		err = author.UpdateById(ctx, artEntity)
	} else {
		id, err = author.Insert(ctx, artEntity)
	}
	if err != nil {
		//tx.Rollback()
		return id, err
	}
	// 操作线上库, 保存数据, 同步过来
	// 考虑到, 此时线上库可能有、可能没有, 你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	err = reader.UpsertV2(ctx, dao.PublishedArticle{Article: artEntity})
	tx.Commit()
	return id, err
}

func (c CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	// 应该先保存到制作库，在保存到线上库
	var (
		id  = art.Id
		err error
	)
	artEntity := c.toEntity(art)
	if id > 0 {
		err = c.authorDAO.UpdateById(ctx, artEntity)
	} else {
		id, err = c.authorDAO.Insert(ctx, artEntity)
	}
	if err != nil {
		return id, err
	}
	// 操作线上库, 保存数据, 同步过来
	// 考虑到, 此时线上库可能有、可能没有, 你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	err = c.readerDAO.Upsert(ctx, artEntity)
	return id, err
}

func (c CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
