# ---------------------------------------------------
# 1. 依赖构建阶段 (Build Stage)
# ---------------------------------------------------
FROM node:22.18.0-alpine AS builder
WORKDIR /home/app

COPY package*.json ./

# 设置镜像源并安装依赖
RUN npm config set registry https://registry.npmmirror.com && \
    npm ci --omit=dev && \
    npm cache clean --force

# ---------------------------------------------------
# 2. 生产运行阶段 (Runner Stage)
# ---------------------------------------------------
FROM node:22.18.0-alpine
WORKDIR /home/app

# 安装 Tini：解决 PID 1 信号转发和僵尸进程问题
RUN apk add --no-cache tini

# 从构建阶段复制 node_modules
COPY --from=builder /home/app/node_modules ./node_modules

# 复制项目文件
COPY . .

EXPOSE 3000

# 使用 Tini 作为入口，确保应用能正常接收停止信号
ENTRYPOINT ["/sbin/tini", "--"]

CMD ["node", "app.js"]
