# 基础镜像
FROM ubuntu:latest
# 把编译后的打包进来这个镜像，放到工作目录 /app 你随便换
WORKDIR /app
COPY webook /app

# CMD 是执行命令
# 最佳
ENTRYPOINT ["/app/webook"]
