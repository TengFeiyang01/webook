package startup

import (
	"github.com/TengFeiyang01/webook/webook/internal/service/oauth2/wechat"
	"github.com/TengFeiyang01/webook/webook/internal/web"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{}
}
