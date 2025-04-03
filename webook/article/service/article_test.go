package service

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"testing"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"github.com/TengFeiyang01/webook/webook/article/repository"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
			repository.ArticleReaderRepository)

		art domain.Article

		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				author := svcmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)
				reader := svcmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
				Status: domain.ArticleStatusPublished,
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "修改并发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				author := svcmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := svcmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 2,
		},
		{
			name: "保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				author := svcmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(0), errors.New("mock error"))
				reader := svcmocks.NewMockArticleReaderRepository(ctrl)
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
				Status: domain.ArticleStatusPublished,
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
		{
			name: "保存到制作库成功，重试到线上库成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				author := svcmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := svcmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(0), errors.New("mock error"))
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(2), nil)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  2,
			wantErr: nil,
		},
		{
			name: "保存到制作库成功，重试全部失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
				repository.ArticleReaderRepository) {
				author := svcmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)
				reader := svcmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Times(3).Return(int64(0), errors.New("mock error"))
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
				Status: domain.ArticleStatusPublished,
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			author, reader := tc.mock(ctrl)
			svc := NewArticleServiceV1(author, reader, &logger.NopLogger{})
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
