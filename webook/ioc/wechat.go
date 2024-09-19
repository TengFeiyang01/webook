package ioc

import (
	"os"
	"webook/webook/internal/service/oauth2/wechat"
	"webook/webook/internal/web"
)

func InitOAuth2WechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID is not set")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("WECHAT_APP_SECRET is not set")
	}
	return wechat.NewService(appId, appSecret)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
