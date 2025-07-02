# 网关服务 Dockerfile
FROM golang:1.24.3-alpine AS builder

# 设置工作目录
WORKDIR /app

# 更换Alpine镜像源为中科大镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装基础依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置Go代理为国内镜像源
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=sum.golang.google.cn

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./services/gateway/cmd/main.go

# 最终镜像
FROM alpine:latest

# 更换Alpine镜像源为中科大镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装CA证书和时区数据
RUN apk --no-cache add ca-certificates curl tzdata && \
    update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/gateway .

# 创建必要的目录
RUN mkdir -p logs configs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动命令
CMD ["./gateway"] 