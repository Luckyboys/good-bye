# 使用 Go 官方镜像作为构建环境
FROM golang:1.24.5-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 安装 make（如果需要）
RUN apk add --no-cache make

# 构建应用（使用 Makefile 以确保版本信息正确传递）
RUN make release

# 使用轻量级的 Alpine 镜像作为运行环境
FROM alpine:3.22

# 安装 ca-certificates 以支持 HTTPS 请求
RUN apk update && \
    apk upgrade && \
    apk add --no-cache bash bash-completion bash-doc && \
    apk add --no-cache ca-certificates curl tzdata && \
    rm -rf /var/cache/apk/* && \
    sh -c 'bash'

# 设置工作目录
WORKDIR /app/

# 从构建阶段复制二进制文件
COPY --from=builder /app/builds/release/good-bye .

# 拷贝静态资源文件
COPY templates /app/templates
COPY static /app/static

# 创建配置文件目录
RUN mkdir -p ./config ./data ./logs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 运行应用
CMD ["./good-bye"]