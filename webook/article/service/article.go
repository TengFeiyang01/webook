package service

import (
	"context"
	"time"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/article/events"
	"github.com/TengFeiyang01/webook/webook/article/repository"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

//go:generate mockgen -source=./art.go -destination=./mocks/service.mock.go -package=svcmocks ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	WithDraw(ctx context.Context, art domain.Article) error
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, id int64, offset int, limit int) ([]domain.Article, error)
	// ListPub 只会取 start 七天内的数据
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository

	// V1
	author   repository.ArticleAuthorRepository
	reader   repository.ArticleReaderRepository
	producer events.Producer
	l        logger.LoggerV1

	ch chan readInfo
}

func (svc *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return svc.repo.ListPub(ctx, start, offset, limit)
}

type readInfo struct {
	aid int64 `json:"aid"`
	uid int64 `json:"uid"`
}

func (svc *articleService) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	art, err := svc.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			// 生产者也可以通过改批量来提高性能
			err := svc.producer.ProduceReadEvent(
				ctx,
				events.ReadEvent{
					// 即便你的消费者要用 art 的数据
					// 让它去查, 你不要在 event 里面带
					Uid: uid,
					Aid: art.Id,
				},
			)
			if err != nil {
				svc.l.Error("failed to send read event")
			}
		}()

		go func() {
			svc.ch <- readInfo{
				aid: art.Id,
				uid: uid,
			}
		}()
	}
	return art, err
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

func NewArticleService(repo repository.ArticleRepository, producer events.Producer, l logger.LoggerV1) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
	}
}

func NewArticleServiceV2(repo repository.ArticleRepository, producer events.Producer, l logger.LoggerV1) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		for {
			uids := make([]int64, 0, 10)
			aids := make([]int64, 0, 10)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			for i := 0; i < 10; i++ {
				select {
				case info, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					uids = append(uids, info.uid)
					aids = append(aids, info.aid)
				case <-ctx.Done():
					break
				}
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			_ = producer.ProduceReadEventV1(ctx, events.ReadEventV1{
				Uid: uids,
				Aid: aids,
			})
		}
	}()
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
		ch:       ch,
	}
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository, reader repository.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
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
