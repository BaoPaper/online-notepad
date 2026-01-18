# 使用Node.js 22.2.0作为基础镜像
FROM node:22.18.0-alpine

# 设置工作目录
WORKDIR /home/app

# 复制package.json和package-lock.json
COPY package*.json ./

# 设置npm镜像
RUN npm config set registry https://registry.npmmirror.com

# 安装依赖
RUN npm install

# 复制项目文件
COPY . .

# 暴露端口
EXPOSE 3000

# 启动应用
CMD ["node", "app.js"]
