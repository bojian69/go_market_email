#!/bin/bash

# AI邮件营销系统部署脚本

set -e

echo "🚀 开始部署 AI邮件营销系统..."

# 检查环境
check_requirements() {
    echo "📋 检查环境要求..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查docker-compose
    if ! command -v docker-compose &> /dev/null; then
        echo "❌ docker-compose 未安装，请先安装 docker-compose"
        exit 1
    fi
    
    echo "✅ 环境检查通过"
}

# 构建前端
build_frontend() {
    echo "🔨 构建前端..."
    cd web
    
    if [ ! -d "node_modules" ]; then
        echo "📦 安装前端依赖..."
        npm install
    fi
    
    echo "🏗️ 构建前端项目..."
    npm run build
    
    cd ..
    echo "✅ 前端构建完成"
}

# 构建后端
build_backend() {
    echo "🔨 构建后端..."
    
    echo "📦 下载Go依赖..."
    go mod tidy
    
    echo "🏗️ 构建Go项目..."
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/cli ./cmd/cli
    
    echo "✅ 后端构建完成"
}

# 构建Docker镜像
build_docker() {
    echo "🐳 构建Docker镜像..."
    
    docker build -t go-market-email:latest .
    
    echo "✅ Docker镜像构建完成"
}

# 部署服务
deploy_services() {
    echo "🚀 部署服务..."
    
    # 创建必要的目录
    mkdir -p logs
    mkdir -p data/mysql
    mkdir -p data/redis
    
    # 复制配置文件
    if [ ! -f "configs/config.local.yaml" ]; then
        echo "📝 创建本地配置文件..."
        cp configs/config.yaml configs/config.local.yaml
        echo "⚠️  请编辑 configs/config.local.yaml 配置文件"
    fi
    
    # 启动服务
    echo "🔄 启动Docker服务..."
    docker-compose up -d
    
    echo "✅ 服务部署完成"
}

# 检查服务状态
check_services() {
    echo "🔍 检查服务状态..."
    
    sleep 10
    
    # 检查容器状态
    if docker-compose ps | grep -q "Up"; then
        echo "✅ 服务运行正常"
        
        echo "📊 服务状态:"
        docker-compose ps
        
        echo ""
        echo "🌐 访问地址:"
        echo "  前端: http://localhost:8080"
        echo "  API:  http://localhost:8080/api/v1"
        echo "  健康检查: http://localhost:8080/health"
        
    else
        echo "❌ 服务启动失败"
        echo "📋 查看日志:"
        docker-compose logs
        exit 1
    fi
}

# 显示使用说明
show_usage() {
    echo ""
    echo "📖 使用说明:"
    echo "  启动服务: docker-compose up -d"
    echo "  停止服务: docker-compose down"
    echo "  查看日志: docker-compose logs -f"
    echo "  重启服务: docker-compose restart"
    echo ""
    echo "🔧 配置文件:"
    echo "  主配置: configs/config.local.yaml"
    echo "  环境变量: .env"
    echo ""
    echo "📁 重要目录:"
    echo "  日志: ./logs/"
    echo "  数据: ./data/"
}

# 主函数
main() {
    case "${1:-all}" in
        "check")
            check_requirements
            ;;
        "frontend")
            build_frontend
            ;;
        "backend")
            build_backend
            ;;
        "docker")
            build_docker
            ;;
        "deploy")
            deploy_services
            ;;
        "status")
            check_services
            ;;
        "all")
            check_requirements
            build_frontend
            build_backend
            build_docker
            deploy_services
            check_services
            show_usage
            ;;
        *)
            echo "用法: $0 {check|frontend|backend|docker|deploy|status|all}"
            echo ""
            echo "  check    - 检查环境要求"
            echo "  frontend - 构建前端"
            echo "  backend  - 构建后端"
            echo "  docker   - 构建Docker镜像"
            echo "  deploy   - 部署服务"
            echo "  status   - 检查服务状态"
            echo "  all      - 执行完整部署流程"
            exit 1
            ;;
    esac
}

main "$@"