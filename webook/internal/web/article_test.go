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
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	svcmocks "webook/webook/internal/service/mocks"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/pkg/ginx"
	"webook/webook/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {

	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
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
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "new title",
					Content: "new content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
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
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish failed"))
				return svc
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
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
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
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), gorm.ErrRecordNotFound)
				return svc
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
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
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
