package art

import (
	"context"
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
)

type GrayScaleArticleServiceClient struct {
	local     artv1.ArticleServiceClient
	remote    artv1.ArticleServiceClient
	threshold *atomicx.Value[int32]
}

func NewGrayScaleArticleServiceClient(local artv1.ArticleServiceClient, remote artv1.ArticleServiceClient) *GrayScaleArticleServiceClient {
	return &GrayScaleArticleServiceClient{local: local, remote: remote, threshold: atomicx.NewValue[int32]()}
}

func (g *GrayScaleArticleServiceClient) UpdateThreshold(threshold int32) {
	g.threshold.Store(threshold)
}

func (g *GrayScaleArticleServiceClient) client() artv1.ArticleServiceClient {
	num := rand.Int31n(100)
	if num <= g.threshold.Load() {
		return g.remote
	}
	return g.local
}

func (g *GrayScaleArticleServiceClient) Save(ctx context.Context, in *artv1.SaveRequest, opts ...grpc.CallOption) (*artv1.SaveResponse, error) {
	return g.client().Save(ctx, in)
}

func (g *GrayScaleArticleServiceClient) WithDraw(ctx context.Context, in *artv1.WithDrawRequest, opts ...grpc.CallOption) (*artv1.WithDrawResponse, error) {
	return g.client().WithDraw(ctx, in)
}

func (g *GrayScaleArticleServiceClient) Publish(ctx context.Context, in *artv1.PublishRequest, opts ...grpc.CallOption) (*artv1.PublishResponse, error) {
	return g.client().Publish(ctx, in)
}

func (g *GrayScaleArticleServiceClient) List(ctx context.Context, in *artv1.ListRequest, opts ...grpc.CallOption) (*artv1.ListResponse, error) {
	return g.client().List(ctx, in)
}

func (g *GrayScaleArticleServiceClient) ListPub(ctx context.Context, in *artv1.ListPubRequest, opts ...grpc.CallOption) (*artv1.ListPubResponse, error) {
	return g.client().ListPub(ctx, in)
}

func (g *GrayScaleArticleServiceClient) GetById(ctx context.Context, in *artv1.GetByIdRequest, opts ...grpc.CallOption) (*artv1.GetByIdResponse, error) {
	return g.client().GetById(ctx, in)
}

func (g *GrayScaleArticleServiceClient) GetPubById(ctx context.Context, in *artv1.GetPubByIdRequest, opts ...grpc.CallOption) (*artv1.GetPubByIdResponse, error) {
	return g.client().GetPubById(ctx, in)
}
