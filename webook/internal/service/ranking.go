package service

import (
	"errors"
	intrv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/intr/v1"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	service2 "github.com/TengFeiyang01/webook/webook/article/service"
	"github.com/TengFeiyang01/webook/webook/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/net/context"
	"math"
	"time"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=./mocks/ranking.mock.go RankingService
type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    service2.ArticleService
	interSvc  intrv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int
	n         int

	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc service2.ArticleService, interSvc intrv1.InteractiveServiceClient) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		interSvc:  interSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2, 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里, 存起来, 塞进去 Redis 里面
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	now := time.Now()
	ddl := now.Add(-time.Hour * 24 * 7)
	type Score struct {
		art   domain.Article
		score float64
	}
	q := queue.NewPriorityQueue[Score](svc.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})

	for {
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})

		// 要去找到对应的点赞数据
		resp, err := svc.interSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz:    "art",
			BizIds: ids,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Intrs) == 0 {
			return nil, errors.New("没有数据")
		}

		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := resp.Intrs[art.Id]

			// 规避负数问题
			score := svc.scoreFunc(art.Utime, intr.LikeCnt+2)

			err = q.Enqueue(Score{
				art:   art,
				score: score,
			})

			if errors.Is(err, queue.ErrOutOfCapacity) {
				// 我要考虑, 我这个 score 在不在前 100 名
				val, _ := q.Dequeue()
				if val.score < score {
					val = Score{
						art:   art,
						score: score,
					}
				}
				_ = q.Enqueue(val)
			}
		}

		// 一批已经处理完了, 问题来了, 我要不要进入下一批？我怎么知道还有没有
		if len(arts) == 0 || len(arts) < svc.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
		// 更新 offset
		offset = offset + len(arts)
	}
	// 最后得出结果
	ql := q.Len()
	res := make([]domain.Article, ql)
	for i := ql - 1; i >= 0; i-- {
		val, _ := q.Dequeue()
		res[i] = val.art
	}
	return res, nil
}
