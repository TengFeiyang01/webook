//go:build manual

package wechat

import (
	"context"
	"os"
	"testing"
)

func Test_service_e2e_VerifyCode(t *testing.T) {
	// TODO: 获取 appid、appSecret hosts修改域名为本地
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID is not set")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("WECHAT_APP_SECRET is not set")
	}
	svc := NewService(appId, appSecret)
	res, err := svc.VerifyCode(context.Background(), "", "state")
	if err != nil {
		return
	}
	t.Log(res)
}
