# 基础镜像
FROM ubuntu:latest

# 把编译后的打包进来这个镜像，放到工作目录 /app（你可以根据实际情况修改路径）
WORKDIR /app

# 复制应用程序文件到容器中
COPY webook /app

# 添加执行权限给应用程序文件
RUN chmod +x /app/webook

# CMD 是执行命令
# 指定应用程序作为入口命令
ENTRYPOINT ["/app/webook"]
