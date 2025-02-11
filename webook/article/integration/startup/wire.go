//go:build wireinject

package startup

import (
	"github.com/google/wire"
	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/article/events"
	"webook/webook/article/repository"
	"webook/webook/article/repository/cache"
	"webook/webook/article/repository/dao"
	"webook/webook/article/service"
	usrrepo "webook/webook/internal/repository"
	usrcache "webook/webook/internal/repository/cache"
	usrdao "webook/webook/internal/repository/dao"
	usrsvc "webook/webook/internal/service"
	"webook/webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB, InitLogger, InitKafka, ioc.NewSyncProducer,
)

var userSvcProviderSet = wire.NewSet(
	usrdao.NewUserDAO,
	usrrepo.NewUserRepository,
	usrsvc.NewUserService,
	usrcache.NewRedisUserCache)

var articlSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	dao.NewGORMArticleDAO,
	service.NewArticleService,
	events.NewKafkaProducer,
	intrv1.NewInteractiveServiceClient,
	cache.NewArticleCache)

func InitArticleHandler() service.ArticleService {
	wire.Build(articlSvcProvider, thirdPartySet, userSvcProviderSet)
	return service.NewArticleService(nil, nil, nil)
}
