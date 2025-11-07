# AI邮件营销系统 Makefile

.PHONY: help build run test clean docker deploy

# 默认目标
help:
	@echo "AI邮件营销系统 - 可用命令:"
	@echo ""
	@echo "  build      - 构建项目"
	@echo "  run        - 运行开发服务器"
	@echo "  test       - 运行测试"
	@echo "  clean      - 清理构建文件"
	@echo "  docker     - 构建Docker镜像"
	@echo "  deploy     - 部署到生产环境"
	@echo "  dev        - 启动开发环境"
	@echo "  lint       - 代码检查"
	@echo "  format     - 格式化代码"

# 构建项目
build:
	@echo "🔨 构建后端..."
	@go mod tidy
	@CGO_ENABLED=0 go build -o bin/server ./cmd/server
	@CGO_ENABLED=0 go build -o bin/cli ./cmd/cli
	@echo "🔨 构建前端..."
	@cd web && npm install && npm run build
	@echo "✅ 构建完成"

# 运行开发服务器
run:
	@echo "🚀 启动开发服务器..."
	@go run cmd/server/main.go -c configs/config.local.yaml

# 运行CLI工具
worker:
	@echo "🔄 启动邮件发送工作进程..."
	@go run cmd/cli/main.go worker -c configs/config.local.yaml

# 运行测试
test:
	@echo "🧪 运行测试..."
	@go test -v ./...
	@cd web && npm test

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	@rm -rf bin/
	@rm -rf web/dist/
	@rm -rf logs/
	@echo "✅ 清理完成"

# 构建Docker镜像
docker:
	@echo "🐳 构建Docker镜像..."
	@docker build -t go-market-email:latest .
	@echo "✅ Docker镜像构建完成"

# 部署到生产环境
deploy:
	@echo "🚀 部署到生产环境..."
	@./scripts/deploy.sh all

# 启动开发环境
dev:
	@echo "🔧 启动开发环境..."
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "前端开发服务器: http://localhost:3000"
	@echo "后端API服务器: http://localhost:8080"

# 代码检查
lint:
	@echo "🔍 代码检查..."
	@golangci-lint run
	@cd web && npm run lint

# 格式化代码
format:
	@echo "✨ 格式化代码..."
	@go fmt ./...
	@cd web && npm run format

# 生成API文档
docs:
	@echo "📚 生成API文档..."
	@swag init -g cmd/server/main.go

# 数据库迁移
migrate:
	@echo "🗄️ 数据库迁移..."
	@go run cmd/migrate/main.go

# 性能测试
benchmark:
	@echo "⚡ 性能测试..."
	@go test -bench=. -benchmem ./...

# 安全检查
security:
	@echo "🔒 安全检查..."
	@gosec ./...

# 依赖更新
update:
	@echo "📦 更新依赖..."
	@go get -u ./...
	@go mod tidy
	@cd web && npm update

# 生成版本信息
version:
	@echo "📋 版本信息:"
	@echo "Git Commit: $(shell git rev-parse HEAD)"
	@echo "Build Time: $(shell date)"
	@echo "Go Version: $(shell go version)"

# 监控日志
logs:
	@echo "📋 查看日志..."
	@tail -f logs/app.log

# 健康检查
health:
	@echo "🏥 健康检查..."
	@curl -f http://localhost:8080/health || echo "服务未运行"