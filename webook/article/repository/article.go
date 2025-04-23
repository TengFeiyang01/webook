package repository

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/article/repository/cache"
	dao2 "github.com/TengFeiyang01/webook/webook/article/repository/dao"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao     dao2.ArticleDAO
	userDAO dao.UserDAO

	// V1 操作两个 DAO
	readerDAO dao2.ReaderDAO
	authorDAO dao2.AuthorDao

	l logger.LoggerV1

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果你要在 repository 层面上操作事务
	// 那么就只能利用 db 开启事务之后，创建基于事务的 DAO
	// 或者，直接去掉 DAO 这一层、在 repository 的实现中，直接操作 db。
	db *gorm.DB

	cache cache.ArticleCache
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	res, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao2.Article) domain.Article {
		return c.toDomain(src)
	}), nil
}

func (c *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被放过去了 OSS 上，就需要前端去读
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 你在这边要组装 user 了
	usr, err := c.userDAO.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res := c.toDomain(art)
	res.Author.Name = usr.NickName

	err = c.cache.SetPub(ctx, id, res)
	if err != nil {
		c.l.Error("提取设置缓存失败",
			logger.Int64("author", art.AuthorId), logger.Error(err))
	}
	return res, nil
}

func (c *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成记得复杂的缓存方案
	// 缓存第一页
	if offset == 0 && limit == 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao2.Article, domain.Article](res, func(idx int, src dao2.Article) domain.Article {
		return c.toDomain(src)
	})
	// 回写缓存的时候，是 Set 还是 Del
	// 如果觉得不太可能有很高并发，你就直接 Set
	// 否则就 Del
	go func() {
		_ = c.cache.SetFirstPage(ctx, uid, data)
		c.l.Error("回写缓存失败", logger.Error(err))
		c.preCache(ctx, data)
	}()
	return data, nil
}

func (c *CachedArticleRepository) toDomain(art dao2.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		author := art.Author.Id
		// 清空缓存
		err := c.cache.DelFirstPage(ctx, art.Id)
		if err != nil {
			c.l.Error("删除缓存失败", logger.Int64("author", author), logger.Int64("artId", art.Id), logger.Error(err))
		}
		// 提前缓存
		err = c.cache.SetPub(ctx, art.Id, art)
		if err != nil {
			c.l.Warn("缓存失败", logger.Int64("author", author), logger.Int64("artId", art.Id), logger.Error(err))
		}
	}
	return id, err
}

// SyncV2 尝试在 repository 层解决事务问题
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	// 利用 tx 来构建DAO
	author := dao2.NewAuthorDAO(tx)
	reader := dao2.NewReaderDAO(tx)
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
	err = reader.UpsertV2(ctx, dao2.PublishedArticleV1{Article: artEntity})
	tx.Commit()
	return id, err
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
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

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err != nil {
		return err
	}
	author := art.Author.Id
	err = c.cache.DelFirstPage(ctx, art.Id)
	if err != nil {
		c.l.Error("删除缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return nil
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {

	id, err := c.dao.Insert(ctx, dao2.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
	if err != nil {
		return 0, err
	}
	author := art.Author.Id
	// 清空缓存
	err = c.cache.DelFirstPage(ctx, art.Id)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("author", author), logger.Error(err))
	}
	return id, nil
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao2.Article {
	return dao2.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		err := c.cache.Set(ctx, arts[0].Id, arts[0])
		if err != nil {
			c.l.Error("提取预加载缓存失败", logger.Error(err))
		}
	}
}

func NewCachedArticleRepository(dao dao2.ArticleDAO, l logger.LoggerV1, usrDAO dao.UserDAO, cache cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		dao:     dao,
		l:       l,
		userDAO: usrDAO,
		cache:   cache,
	}
}
