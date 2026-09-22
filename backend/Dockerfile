# 多阶段构建：先编译二进制，再放入极简运行时镜像
FROM golang:1.27-alpine AS builder

# 国内构建可解除下行注释启用镜像加速
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOSUMDB=off

WORKDIR /src

# 缓存依赖层
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静态编译减少运行时依赖
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/prompthub-backend \
    ./cmd/server

# 运行时镜像
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /out/prompthub-backend /app/prompthub-backend
COPY --from=builder /src/configs /app/configs

USER app

EXPOSE 8080

ENTRYPOINT ["/app/prompthub-backend"]
