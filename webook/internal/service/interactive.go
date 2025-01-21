package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=./mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(c context.Context, biz string, id int64, uid int64) error
	CancelLike(c context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, id, uid)
		return er
	})

	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, id, uid)
		return er
	})
	return intr, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) Like(c context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLike(c, biz, id, uid)
}

func (i *interactiveService) CancelLike(c context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(c, biz, id, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}
