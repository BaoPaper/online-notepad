# ---------------------------------------------------
# 1. Go 构建阶段 (Build Stage)
# ---------------------------------------------------
FROM golang:1.25.8-alpine AS builder
WORKDIR /home/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /simplenote .

# ---------------------------------------------------
# 2. 生产运行阶段 (Runner Stage)
# ---------------------------------------------------
FROM alpine:3.22
WORKDIR /home/app

# 安装 Tini：解决 PID 1 信号转发和僵尸进程问题
RUN apk add --no-cache ca-certificates tini

# 从构建阶段复制 Go 可执行文件
COPY --from=builder /simplenote ./simplenote

# 复制运行时资源
COPY public ./public
COPY views ./views

EXPOSE 3000

# 使用 Tini 作为入口，确保应用能正常接收停止信号
ENTRYPOINT ["/sbin/tini", "--"]

CMD ["./simplenote"]
