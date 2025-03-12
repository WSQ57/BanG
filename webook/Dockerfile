# 基础镜像
FROM ubuntu:20.04

# 把编译后的打包进镜像，工作目录设置为app
COPY webook /app/webook
WORKDIR /app
# CMD是提供默认指令 ENTRYPOINT最佳
ENTRYPOINT ["/app/webook"]
