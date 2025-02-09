package repository

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/slice"
	"webook/webook/interactive/domain"
	"webook/webook/interactive/repository/cache"
	"webook/webook/interactive/repository/dao"
	"webook/webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context,
	bizs []string, bizIds []int64) error {
	// 要不要检测 bizs 和 ids 的长度是否相等
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	// 你也要批量的去修改 redis, 所以就要去改 lua 脚本
	// c.cache.IncrReadCntIfPresent()
	// TODO 新的 lua 脚本 / pipeline
	return nil
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO,
	l logger.LoggerV1,
	cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache, l: l}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, id)
	if err == nil {
		res := c.toDomain(ie)
		err = c.cache.Set(ctx, biz, id, res)
		if err != nil {
			c.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", id),
				logger.Error(err))
		}
		return res, nil
	}
	return intr, err
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context,
	biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context,
	biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context,
	biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	// 这个地方要不要考虑缓存
	// 一个东西要不要缓存，就看用户会不会频繁访问
	return c.cache.IncrCollectCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	err = c.cache.IncrLikeCntIfPresent(ctx, biz, id)
	if err != nil {
		return err
	}
	err = c.cache.IncrRankingIfPresent(ctx, biz, id)
	if errors.Is(err, cache.RankingUpdateErr) {
		val, err := c.dao.Get(ctx, biz, id)
		if err != nil {
			return err
		}
		return c.cache.SetRankingScore(ctx, biz, id, val.LikeCnt)
	}
	return err
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 你要更新缓存了
	// 部分失败问题 —— 数据不一致
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:        ie.Biz,
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
