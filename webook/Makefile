.PHONY: docker
docker:
	@#if exist webook (del webook) else (echo File not found & exit w)
	@#set GOARCH=arm
	@#set GOOS=linux
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rmi -f sparrow/webook:v0.0.1
	@docker build -t sparrow/webook:v0.0.1 .
