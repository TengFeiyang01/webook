package startup

import (
	"webook/webook/internal/service/oauth2/wechat"
	"webook/webook/internal/web"
	"webook/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{}
}
