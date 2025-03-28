package art

import (
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/article/service"
)

type ArticleServiceAdapter struct {
	svc service.ArticleService
}

func NewArticleServiceAdapter(svc service.ArticleService) *ArticleServiceAdapter {
	return &ArticleServiceAdapter{svc: svc}
}

func (a *ArticleServiceAdapter) Save(ctx context.Context, in *artv1.SaveRequest, opts ...grpc.CallOption) (*artv1.SaveResponse, error) {
	id, err := a.svc.Save(ctx, a.toPB(in.GetArt()))
	return &artv1.SaveResponse{Id: id}, err
}

func (a *ArticleServiceAdapter) WithDraw(ctx context.Context, in *artv1.WithDrawRequest, opts ...grpc.CallOption) (*artv1.WithDrawResponse, error) {
	err := a.svc.WithDraw(ctx, a.toPB(in.GetArt()))
	return &artv1.WithDrawResponse{}, err
}

func (a *ArticleServiceAdapter) Publish(ctx context.Context, in *artv1.PublishRequest, opts ...grpc.CallOption) (*artv1.PublishResponse, error) {
	id, err := a.svc.Publish(ctx, a.toPB(in.GetArt()))
	return &artv1.PublishResponse{Id: id}, err
}

func (a *ArticleServiceAdapter) List(ctx context.Context, in *artv1.ListRequest, opts ...grpc.CallOption) (*artv1.ListResponse, error) {
	arts, err := a.svc.List(ctx, in.GetId(), int(in.GetOffset()), int(in.GetLimit()))
	return &artv1.ListResponse{Arts: slice.Map(arts, func(idx int, src domain.Article) *artv1.Article {
		return a.toDTO(arts[idx])
	})}, err
}

func (a *ArticleServiceAdapter) ListPub(ctx context.Context, in *artv1.ListPubRequest, opts ...grpc.CallOption) (*artv1.ListPubResponse, error) {
	arts, err := a.svc.ListPub(ctx, time.UnixMilli(in.GetTimestamp()), int(in.GetOffset()), int(in.GetLimit()))
	return &artv1.ListPubResponse{Arts: slice.Map(arts, func(idx int, src domain.Article) *artv1.Article {
		return a.toDTO(arts[idx])
	})}, err
}

func (a *ArticleServiceAdapter) GetById(ctx context.Context, in *artv1.GetByIdRequest, opts ...grpc.CallOption) (*artv1.GetByIdResponse, error) {
	art, err := a.svc.GetById(ctx, in.GetId())
	return &artv1.GetByIdResponse{Art: a.toDTO(art)}, err
}

func (a *ArticleServiceAdapter) GetPubById(ctx context.Context, in *artv1.GetPubByIdRequest, opts ...grpc.CallOption) (*artv1.GetPubByIdResponse, error) {
	art, err := a.svc.GetById(ctx, in.GetId())
	return &artv1.GetPubByIdResponse{Art: a.toDTO(art)}, err
}

func (a *ArticleServiceAdapter) toDTO(art domain.Article) *artv1.Article {
	return &artv1.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  uint32(art.Status),
		Author: &artv1.Author{
			Id:   art.Author.Id,
			Name: art.Author.Name,
		},
		Ctime: art.Ctime.UnixMilli(),
		Utime: art.Utime.UnixMilli(),
	}
}

func (a *ArticleServiceAdapter) toPB(art *artv1.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id:   art.Author.Id,
			Name: art.Author.Name,
		},
		Status: domain.ArticleStatus(uint8(art.Status)),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}
