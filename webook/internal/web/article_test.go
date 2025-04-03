package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/article/service"
	service2 "github.com/TengFeiyang01/webook/webook/interactive/service"
	svcmocks "github.com/TengFeiyang01/webook/webook/internal/service/mocks"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	loggermocks "github.com/TengFeiyang01/webook/webook/pkg/logger/mocks"
)

func TestArticleHandler_Publish(t *testing.T) {

	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService)

		reqBody string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc, l, interSvc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "已有帖子且发表成功",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "new title",
					Content: "new content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc, l, interSvc
			},
			reqBody: `
{
	"id":1,
	"title":"new title",
	"content":"new content"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "publish失败",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish failed"))
				return svc, l, interSvc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "输入有误、Bind返回错误",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				return svc, l, interSvc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my con
}
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "找不到User",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, logger.LoggerV1, service2.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				l := loggermocks.NewMockLoggerV1(ctrl)
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), gorm.ErrRecordNotFound)
				return svc, l, interSvc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: http.StatusUnauthorized,
				Msg:  "找不到用户",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.UserClaims{
					Uid: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))

			// 这里就可以继续使用 req 了
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes ginx.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
