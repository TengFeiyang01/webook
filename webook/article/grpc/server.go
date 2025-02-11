package grpc

import (
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
	artv1 "webook/webook/api/proto/gen/article/v1"
	"webook/webook/article/domain"
	"webook/webook/article/service"
)

type ArticleServiceServer struct {
	svc service.ArticleService
	artv1.UnimplementedArticleServiceServer
}

func NewArticleServiceServer(svc service.ArticleService) *ArticleServiceServer {
	return &ArticleServiceServer{svc: svc}
}

func (a *ArticleServiceServer) Register(server *grpc.Server) {
	artv1.RegisterArticleServiceServer(server, a)
}

func (a *ArticleServiceServer) Save(ctx context.Context, request *artv1.SaveRequest) (*artv1.SaveResponse, error) {
	id, err := a.svc.Save(ctx, a.toPB(request.GetArt()))
	return &artv1.SaveResponse{Id: id}, err
}

func (a *ArticleServiceServer) WithDraw(ctx context.Context, request *artv1.WithDrawRequest) (*artv1.WithDrawResponse, error) {
	err := a.svc.WithDraw(ctx, a.toPB(request.GetArt()))
	return &artv1.WithDrawResponse{}, err
}

func (a *ArticleServiceServer) Publish(ctx context.Context, request *artv1.PublishRequest) (*artv1.PublishResponse, error) {
	id, err := a.svc.Publish(ctx, a.toPB(request.GetArt()))
	return &artv1.PublishResponse{Id: id}, err
}

func (a *ArticleServiceServer) List(ctx context.Context, request *artv1.ListRequest) (*artv1.ListResponse, error) {
	arts, err := a.svc.List(ctx, request.GetId(), int(request.GetOffset()), int(request.GetLimit()))
	return &artv1.ListResponse{
		Arts: slice.Map(arts, func(idx int, src domain.Article) *artv1.Article {
			return a.toDTO(arts[idx])
		}),
	}, err
}

func (a *ArticleServiceServer) ListPub(ctx context.Context, request *artv1.ListPubRequest) (*artv1.ListPubResponse, error) {
	arts, err := a.svc.ListPub(ctx, time.UnixMilli(request.GetTimestamp()), int(request.GetOffset()), int(request.GetLimit()))
	return &artv1.ListPubResponse{
		Arts: slice.Map(arts, func(idx int, src domain.Article) *artv1.Article {
			return a.toDTO(arts[idx])
		}),
	}, err
}

func (a *ArticleServiceServer) GetById(ctx context.Context, request *artv1.GetByIdRequest) (*artv1.GetByIdResponse, error) {
	art, err := a.svc.GetById(ctx, request.GetId())
	return &artv1.GetByIdResponse{Art: a.toDTO(art)}, err
}

func (a *ArticleServiceServer) GetPubById(ctx context.Context, request *artv1.GetPubByIdRequest) (*artv1.GetPubByIdResponse, error) {
	art, err := a.svc.GetPublishedById(ctx, request.GetId(), request.GetUid())
	return &artv1.GetPubByIdResponse{Art: a.toDTO(art)}, err
}

func (a *ArticleServiceServer) toDTO(art domain.Article) *artv1.Article {
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

func (a *ArticleServiceServer) toPB(art *artv1.Article) domain.Article {
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
