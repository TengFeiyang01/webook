package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
	"github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	cachemocks "github.com/TengFeiyang01/webook/webook/internal/repository/cache/mocks"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	daomocks "github.com/TengFeiyang01/webook/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	// 你要去掉毫秒之外的部分
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		ctx context.Context
		id  int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，但是查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				// 缓存未命中，查了缓存，但是没结果
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					ID: 123,
					Email: sql.NullString{
						String: "123@qq,com",
						Valid:  true,
					},
					Password: "this is password",
					Phone: sql.NullString{
						String: "12312312312",
						Valid:  true,
					},
					Ctime: now.UnixMilli(),
					Utime: now.UnixMilli(),
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					ID:       123,
					Email:    "123@qq,com",
					Password: "this is password",
					Phone:    "12312312312",
					Ctime:    now,
				}).Return(nil)

				return d, c

			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				ID:       123,
				Email:    "123@qq,com",
				Password: "this is password",
				Phone:    "12312312312",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					ID:       123,
					Email:    "123@qq,com",
					Password: "this is password",
					Phone:    "12312312312",
					Ctime:    now,
				}, nil)
				d := daomocks.NewMockUserDAO(ctrl)
				return d, c

			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				ID:       123,
				Email:    "123@qq,com",
				Password: "this is password",
				Phone:    "12312312312",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				// 缓存未命中，查了缓存，但是没结果
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, errors.New("mock error"))
				return d, c

			},
			ctx:      context.Background(),
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
			time.Sleep(time.Second) // 因为是异步 需要等待执行完毕
		})
	}
}
