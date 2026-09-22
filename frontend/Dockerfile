# 多阶段构建：先构建静态资源，再交给 Nginx 提供

FROM node:22-alpine AS builder

WORKDIR /src

# 国内构建可解除下行注释启用镜像加速
# RUN npm config set registry https://registry.npmmirror.com

COPY package*.json ./
RUN npm install --no-audit --no-fund

COPY . .
RUN npm run build

# 运行时镜像：Nginx 提供静态文件 + 反向代理 /api 到后端
FROM nginx:1.27-alpine

# 移除默认配置
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/app.conf
COPY --from=builder /src/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -q -O - http://127.0.0.1/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
