// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/TengFeiyang01/webook/webook/interactive/grpc"
	"github.com/TengFeiyang01/webook/webook/interactive/repository"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/cache"
	"github.com/TengFeiyang01/webook/webook/interactive/repository/dao"
	"github.com/TengFeiyang01/webook/webook/interactive/service"
)

// Injectors from wire.go:

func InitInteractiveService() service.InteractiveService {
	gormDB := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	loggerV1 := InitLogger()
	cmdable := InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	return interactiveService
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	gormDB := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	loggerV1 := InitLogger()
	cmdable := InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	return interactiveServiceServer
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitRedis, InitDB, InitLogger,
	NewSyncProducer,
	InitKafka,
)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO, service.NewInteractiveService, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository)
