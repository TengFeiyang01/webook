package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/webook/internal/web"
	"webook/webook/ioc"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := InitWebUser()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string

		// 你要考虑准备数据，以及验证数据
		before func(*testing.T)

		// 验证并且删除数据
		after func(*testing.T)

		phone string

		reqBody string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 不需要，也就是 redis 里面什么数据也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 你要清理数据
				// phone_code:%s:%s
				val, err := rdb.GetDel(ctx, "phone_code:login:13213130303").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			reqBody: `
{
	"phone": "13213130303"
}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, err := rdb.Set(ctx, "phone_code:login:13213130303", "123123",
					time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13213130303").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码还是原来的值，未被覆盖
				assert.Equal(t, "123123", val)
			},
			reqBody: `
{
	"phone": "13213130303"
}`,
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送太频繁, 请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, err := rdb.Set(ctx, "phone_code:login:13213130303", "123123",
					0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13213130303").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码还是原来的值，未被覆盖
				assert.Equal(t, "123123", val)
			},
			reqBody: `
{
	"phone": "13213130303"
}`,
			wantCode: http.StatusInternalServerError,
			wantBody: web.Result{
				Msg: "系统错误",
			},
		},
		{
			name: "手机号码为空",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
{
	"phone": ""
}`,
			wantCode: http.StatusUnauthorized,
			wantBody: web.Result{
				Msg: "手机号格式错误",
			},
		},
		{
			name: "数据格式有误",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
{
	"phone": "
}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tc.after(t)
		})
	}
}
