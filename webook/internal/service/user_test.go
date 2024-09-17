package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
	repomocks "webook/webook/internal/repository/mocks"
)

func Test_userService_Login(t *testing.T) {
	// 公共时间
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 输入
		ctx      context.Context
		email    string
		password string

		// 输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$MdaklX2GEIz8llHB7DwLlepiBuMYU53FeXjlH7Hyl2VE85F5ZV0GO",
						Phone:    "12312312312",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$MdaklX2GEIz8llHB7DwLlepiBuMYU53FeXjlH7Hyl2VE85F5ZV0GO",
				Phone:    "12312312312",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock db error"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db error"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, bcrypt.ErrMismatchedHashAndPassword)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  bcrypt.ErrMismatchedHashAndPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 具体的测试代码
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			u, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#world123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res), err)
	}
}
