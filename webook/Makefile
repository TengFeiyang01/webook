.PHONY: docker
docker:
	# 把上次编译的东西删掉
	@rm webook || true
	@docker rmi -f ytf0609/webook:v0.0.1
	# 指定编译成在 ARM 架构的 linux 操作系统上运行的可执行文件，
	# 名字叫做 webook
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	# 这里你可以随便改这个标签，记得对应的 k8s 部署里面也要改
	@docker build -t ytf0609/webook:v0.0.1 .