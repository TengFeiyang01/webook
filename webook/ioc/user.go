package ioc

import (
	"go.uber.org/zap"
	"webook/webook/internal/repository"
	"webook/webook/internal/service"
	"webook/webook/pkg/logger"
)

func InitUserHandler(repo repository.UserRepository) service.UserService {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(l))
}
