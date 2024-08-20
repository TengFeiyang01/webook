# 基础镜像
FROM ubuntu:20.04

# 把编译后的 webook 打包近来这个镜像，放到 /app 里面
COPY webook /app/webook
# 设置工作目录
WORKDIR /app
# CMD 是执行命令
# 最佳
ENTRYPOINT ["/app/webook"]
