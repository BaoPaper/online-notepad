# ---------------------------------------------------
# 1. 依赖构建阶段 (Build Stage)
# ---------------------------------------------------
FROM node:22.18.0-alpine AS builder
WORKDIR /app

COPY package*.json ./

# 设置镜像源并安装依赖
RUN npm config set registry https://registry.npmmirror.com && \
    npm ci --omit=dev && \
    npm cache clean --force

# ---------------------------------------------------
# 2. 生产运行阶段 (Runner Stage)
# ---------------------------------------------------
FROM node:22.18.0-alpine
WORKDIR /app

# 安装 Tini：解决 PID 1 信号转发和僵尸进程问题
RUN apk add --no-cache tini

# 从构建阶段复制 node_modules，保持 root 权限以避免升级后的权限报错
COPY --from=builder /app/node_modules ./node_modules
COPY . .

EXPOSE 3000

# 使用 Tini 作为入口，确保应用能正常接收停止信号
ENTRYPOINT ["/sbin/tini", "--"]

CMD ["node", "app.js"]
