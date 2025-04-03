//go:build wireinject

package startup

import (
	"github.com/google/wire"
	intrv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/intr/v1"
	"github.com/TengFeiyang01/webook/webook/article/events"
	"github.com/TengFeiyang01/webook/webook/article/repository"
	"github.com/TengFeiyang01/webook/webook/article/repository/cache"
	"github.com/TengFeiyang01/webook/webook/article/repository/dao"
	"github.com/TengFeiyang01/webook/webook/article/service"
	usrrepo "github.com/TengFeiyang01/webook/webook/internal/repository"
	usrcache "github.com/TengFeiyang01/webook/webook/internal/repository/cache"
	usrdao "github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	usrsvc "github.com/TengFeiyang01/webook/webook/internal/service"
	"github.com/TengFeiyang01/webook/webook/ioc"
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
