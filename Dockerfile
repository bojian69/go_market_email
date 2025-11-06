# 多阶段构建
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --only=production

COPY web/ ./
RUN npm run build

# Go构建阶段
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cli ./cmd/cli

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# 复制构建产物
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/cli .
COPY --from=frontend-builder /app/web/dist ./web/dist
COPY configs/ ./configs/

# 创建日志目录
RUN mkdir -p logs

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]