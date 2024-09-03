.PHONY: mock
mock:
	@mockgen -source=webook/internal/service/types.go -package=svcmocks -destination=webook/internal/service/mocks/service.mock.go
	@mockgen -source=webook/internal/repository/types.go -package=repomocks -destination=webook/internal/repository/mocks/repository.mock.go
	@go mod tidy