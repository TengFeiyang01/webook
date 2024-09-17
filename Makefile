.PHONY: mock
mock:
	@mockgen -source=webook/internal/service/types.go -package=svcmocks -destination=webook/internal/service/mocks/service.mock.go
	@mockgen -source=webook/internal/repository/types.go -package=repomocks -destination=webook/internal/repository/mocks/repository.mock.go
	@mockgen -source=webook/internal/repository/cache/types.go -package=cachemocks -destination=webook/internal/repository/cache/mocks/cache.mock.go
	@mockgen -source=webook/internal/repository/dao/types.go -package=daomocks -destination=webook/internal/repository/dao/mocks/dao.mock.go
	@mockgen -package=redismocks -destination=webook/internal/repository/cache/redismocks/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
	@go mod tidy