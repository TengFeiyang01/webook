package service

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"testing"
	"time"
	intrv1 "webook/webook/api/proto/gen/intr/v1"
	intrv1mocks "webook/webook/api/proto/gen/intr/v1/mocks"
	"webook/webook/article/domain"
	service2 "webook/webook/article/service"
	svcmocks "webook/webook/internal/service/mocks"
)

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (service2.ArticleService, intrv1.InteractiveServiceClient)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			// 怎么模拟我的数据
			mock: func(ctrl *gomock.Controller) (service2.ArticleService, intrv1.InteractiveServiceClient) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				interSvc := intrv1mocks.NewMockInteractiveServiceClient(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Utime: now, Ctime: now},
						{Id: 2, Utime: now, Ctime: now},
						{Id: 3, Utime: now, Ctime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).
					Return([]domain.Article{}, nil)

				interSvc.EXPECT().GetByIds(gomock.Any(), &intrv1.GetByIdsRequest{
					Biz:    "art",
					BizIds: []int64{1, 2, 3},
				}).
					Return(&intrv1.GetByIdsResponse{Intrs: map[int64]*intrv1.Interactive{
						1: {BizId: 1, LikeCnt: 1},
						2: {BizId: 2, LikeCnt: 2},
						3: {BizId: 3, LikeCnt: 3},
					}}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), &intrv1.GetByIdsRequest{
					Biz:    "art",
					BizIds: []int64{},
				}).
					Return(&intrv1.GetByIdsResponse{Intrs: map[int64]*intrv1.Interactive{}}, nil)
				return artSvc, interSvc
			},

			wantErr: nil,
			wantArts: []domain.Article{
				{Id: 3, Utime: now, Ctime: now},
				{Id: 2, Utime: now, Ctime: now},
				{Id: 1, Utime: now, Ctime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			artSvc, interSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(artSvc, interSvc).(*BatchRankingService)
			svc.n = 3
			svc.batchSize = 3
			svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
