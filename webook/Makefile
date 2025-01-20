.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .

	@docker rmi -f sparrow/webook:v0.0.1
	@docker build -t sparrow/webook:v0.0.1 .
