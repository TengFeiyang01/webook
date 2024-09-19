package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache/redismocks"
	"webook/webook/internal/service"
	svcmocks "webook/webook/internal/service/mocks"
	ijwt "webook/webook/internal/web/jwt"
	jwtmocks "webook/webook/internal/web/jwt/mocks"
)

func TestEncrypt(t *testing.T) {
	password := []byte("hello#123")
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(hash))
	err = bcrypt.CompareHashAndPassword(hash, password)
	assert.NoError(t, err)
}

func TestUserHandler_SignUp(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler)
		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "2196442691@qq.com",
					Password: "123#@qqcom",
				}).Return(nil)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusOK,
			wantBody: `hello 注册成功`,
		},
		{
			name: "参数不对, bind 失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "12
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `系统错误`,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusUnauthorized,
			wantBody: `你的邮箱格式不对`,
		},
		{
			name: "两次输入的密码不一致",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "1233#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusUnauthorized,
			wantBody: `两次输入的密码不一致`,
		},
		{
			name: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "1233#@q",
    "confirm_password": "1233#@q"
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `密码必须包含数字、特殊字符，并且长度不能小于 8 位`,
		},
		{
			name: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "1233#@q",
    "confirm_password": "1233#@q"
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `密码必须包含数字、特殊字符，并且长度不能小于 8 位`,
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "2196442691@qq.com",
					Password: "123#@qqcom",
				}).Return(service.ErrUserDuplicate)
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusOK,
			wantBody: `邮箱冲突`,
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "2196442691@qq.com",
					Password: "123#@qqcom",
				}).Return(errors.New("any"))
				return userSvc, nil, cmd, jwtHdl
			},
			reqBody: `
{
    "email": "2196442691@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusInternalServerError,
			wantBody: `系统异常`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()

			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))

			// 这里就可以继续使用 req 了
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			// 这就是 HTTP 请求进去 GIN 框架的入口
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回 resp
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	// 先创建一个控制 mock 的控制器
	ctrl := gomock.NewController(t)
	// 每个测试结束都要调用 Finish
	// 然后 mock 就会去验证你的测试流程是否符合预期
	defer ctrl.Finish()

	svc := svcmocks.NewMockUserService(ctrl)
	// 开启一个个测试调用
	// 预期第一个是 Signup 的调用
	// 模拟的 条件是

	svc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
		Return(errors.New("mock error"))

	err := svc.SignUp(context.Background(), domain.User{
		Email: "test@test.com",
	})
	t.Log(err)
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler)

		reqBody  string
		wantCode int
		wantBody Result
	}{
		{
			name: "验证码校验成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				svc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), biz, gomock.Any(), "123456").Return(true, nil)
				svc.EXPECT().FindOrCreate(gomock.Any(), gomock.Any()).Return(domain.User{
					Phone: "12345678901",
				}, nil)
				return svc, codeSvc, cmd, jwtHdl
			},
			wantCode: http.StatusOK,
			reqBody: `
{
	"phone": "12345678901",
	"code": "123456"
}
`,
			wantBody: Result{
				Code: http.StatusOK,
				Msg:  "验证码校验成功",
			},
		},
		{
			name: "参数不对, bind 失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				svc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return svc, codeSvc, cmd, jwtHdl
			},
			wantCode: http.StatusBadRequest,
			reqBody: `
{
	"phone": 5678901",
	"code": "123456"
}
`,
			wantBody: Result{
				Code: http.StatusBadRequest,
				Msg:  "bind失败",
			},
		},
		{
			name: "验证码校验出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				svc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), biz, gomock.Any(), "123456").Return(false, errors.New("mock error"))
				return svc, codeSvc, cmd, jwtHdl
			},
			wantCode: http.StatusInternalServerError,
			reqBody: `
{
	"phone": "12345678901",
	"code": "123456"
}
`,
			wantBody: Result{
				Code: http.StatusInternalServerError,
				Msg:  "系统错误",
			},
		},
		{
			name: "验证码不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				svc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), biz, gomock.Any(), "123456").Return(false, nil)
				return svc, codeSvc, cmd, jwtHdl
			},
			wantCode: http.StatusUnauthorized,
			reqBody: `
{
	"phone": "12345678901",
	"code": "123456"
}
`,
			wantBody: Result{
				Code: http.StatusUnauthorized,
				Msg:  "验证码不正确",
			},
		},
		{
			name: "用户不存在/创建用户失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, redis.Cmdable, ijwt.Handler) {
				cmd := redismocks.NewMockCmdable(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				svc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), biz, gomock.Any(), "123456").Return(true, nil)
				svc.EXPECT().FindOrCreate(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("mock error"))
				return svc, codeSvc, cmd, jwtHdl
			},
			wantCode: http.StatusInternalServerError,
			reqBody: `
{
	"phone": "12345678901",
	"code": "123456"
}
`,
			wantBody: Result{
				Code: http.StatusInternalServerError,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			var res Result
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
