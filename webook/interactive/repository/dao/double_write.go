package dao

import (
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/net/context"
)

const (
	patternSrcOnly  = "patternSrcOnly"
	patternSrcFirst = "patternSrcFirst"
	patternDstOnly  = "patternDstOnly"
	patternDstFirst = "DST_ONLY"
)

var errUnknownPattern = errors.New("未知的模式")

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
}

func NewDoubleWriteDAO(src InteractiveDAO, dst InteractiveDAO) InteractiveDAO {
	return &DoubleWriteDAO{src: src, dst: dst, pattern: atomicx.NewValueOf(patternSrcOnly)}
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case patternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	case patternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	case patternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		return d.src.IncrReadCnt(ctx, biz, bizId)
	}
	return nil
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.InsertLikeInfo(ctx, biz, id, uid)
	case patternSrcFirst:
		err := d.src.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		return d.dst.InsertLikeInfo(ctx, biz, id, uid)
	case patternDstOnly:
		return d.dst.InsertLikeInfo(ctx, biz, id, uid)
	case patternDstFirst:
		err := d.dst.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		return d.src.InsertLikeInfo(ctx, biz, id, uid)
	}
	return nil
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, id, uid)
	case patternSrcFirst:
		err := d.src.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		return d.dst.DeleteLikeInfo(ctx, biz, id, uid)
	case patternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, id, uid)
	case patternDstFirst:
		err := d.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		return d.src.DeleteLikeInfo(ctx, biz, id, uid)
	}
	return nil
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.InsertCollectionBiz(ctx, cb)
	case patternSrcFirst:
		err := d.src.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		return d.dst.InsertCollectionBiz(ctx, cb)
	case patternDstOnly:
		return d.dst.InsertCollectionBiz(ctx, cb)
	case patternDstFirst:
		err := d.dst.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		return d.src.InsertCollectionBiz(ctx, cb)
	}
	return nil
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	switch d.pattern.Load() {
	case patternSrcFirst, patternSrcOnly:
		return d.src.GetLikeInfo(ctx, biz, id, uid)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetLikeInfo(ctx, biz, id, uid)
	default:
		return UserLikeBiz{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	switch d.pattern.Load() {
	case patternSrcFirst, patternSrcOnly:
		return d.src.GetCollectInfo(ctx, biz, id, uid)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetCollectInfo(ctx, biz, id, uid)
	default:
		return UserCollectionBiz{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	switch d.pattern.Load() {
	case patternSrcFirst, patternSrcOnly:
		return d.src.Get(ctx, biz, id)
	case patternDstFirst, patternDstOnly:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.BatchIncrReadCnt(ctx, bizs, ids)
	case patternSrcFirst:
		err := d.src.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			return err
		}
		return d.dst.BatchIncrReadCnt(ctx, bizs, ids)
	case patternDstOnly:
		return d.dst.BatchIncrReadCnt(ctx, bizs, ids)
	case patternDstFirst:
		err := d.dst.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			return err
		}
		return d.src.BatchIncrReadCnt(ctx, bizs, ids)
	}
	return nil
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.GetByIds(ctx, biz, ids)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetByIds(ctx, biz, ids)
	default:
		return nil, errUnknownPattern
	}
}
