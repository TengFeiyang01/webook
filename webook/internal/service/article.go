package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/article"
	"webook/webook/pkg/logger"
)

type articleService struct {
	repo article.ArticleRepository

	// V1
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository

	l logger.LoggerV1
}

func (svc *articleService) GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error) {
	return svc.repo.GetPublishedById(ctx, id)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetByID(ctx, id)
}

func (svc *articleService) List(ctx context.Context, id int64, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, id, offset, limit)
}

func (svc *articleService) WithDraw(ctx context.Context, art domain.Article) error {
	return svc.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	// 制作库
	//id, err := svc.repo.Save(ctx, art)
	// 线上库呢?
	//panic("implement me")
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

func (svc *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	// 制作库
	//id, err := svc.repo.Save(ctx, art)
	// 线上库呢?
	art.Status = domain.ArticleStatusPublished
	var (
		id  = art.Id
		err error
	)
	if art.Id == 0 {
		id, err = svc.author.Create(ctx, art)
	} else {
		err = svc.author.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		id, err = svc.reader.Save(ctx, art)
		if err == nil {
			break
		}
		svc.l.Error("部分失败，保存到线上库失败",
			logger.Int64("art_id", id),
			logger.Error(err))
	}
	if err != nil {
		svc.l.Error("部分失败，重试彻底失败",
			logger.Int64("art_id", id),
			logger.Error(err))
		// 接入你的告警系统，手工处理一下
		// 走异步，我直接保存到本地文件
		// 走 Canal
		// 打 MQ
	}
	return id, err
}

func NewArticleService(repo article.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnPublished
	println(art.Id)
	if art.Id > 0 {
		err := svc.repo.Update(ctx, art)
		return art.Id, err
	}
	return svc.repo.Create(ctx, art)
}
