//go:build wireinject

// 让wire来注入这里的代码

package wire

import (
	"github.com/google/wire"
	"webook/wire/repository"
	"webook/wire/repository/dao"
)

func InitRepository() *repository.UserRepository {
	// 这个方法传入各个组件的初始化方法
	wire.Build(dao.NewUserDAO, repository.NewUserRepository, InitDB)
	return new(repository.UserRepository)
}
