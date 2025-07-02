# 管理API服务 Dockerfile
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# 更换Alpine镜像源为中科大镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk add --no-cache git ca-certificates tzdata

# 设置Go代理为国内镜像源
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin-api ./services/admin-api/cmd/main.go

FROM alpine:latest

# 更换Alpine镜像源为中科大镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk --no-cache add ca-certificates curl tzdata && update-ca-certificates

ENV TZ=Asia/Shanghai
RUN addgroup -g 1000 -S appgroup && adduser -u 1000 -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/admin-api .
RUN mkdir -p logs configs && chown -R appuser:appgroup /app

USER appuser
EXPOSE 8082

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

CMD ["./admin-api"] 