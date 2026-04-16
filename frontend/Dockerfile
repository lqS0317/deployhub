# Stage 1: 安装依赖
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --frozen-lockfile

# Stage 2: 构建（线上通过 Nginx 同域代理，API_URL 为空即走相对路径）
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL=""
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN corepack enable && pnpm build

# Stage 3: 运行阶段
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN mkdir -p public
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
