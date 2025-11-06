# Go Market Email - AI邮件营销系统

基于Go、Vue 3和AI的智能邮件营销系统，支持模板管理、数据导入、AI内容生成和批量发送。

## 功能特性

- 📧 **邮件模板管理**：支持变量提取、版本控制
- 📊 **数据导入**：Excel/CSV文件导入、SQL查询、手动输入
- 🤖 **AI集成**：OpenAI GPT-4、自定义API支持
- 🚀 **批量发送**：队列处理、频率控制、失败重试
- 📈 **实时监控**：发送统计、进度跟踪、可视化面板
- 🔧 **任务管理**：暂停/恢复、CLI调度工具
- 🔒 **安全认证**：Token认证、权限控制

## 技术栈

### 后端
- **Go 1.21** - 主要开发语言
- **Gin** - Web框架
- **GORM** - ORM框架
- **Redis** - 缓存和队列
- **MySQL 8.0** - 数据存储
- **Zap** - 日志管理
- **Viper** - 配置管理
- **Cobra** - CLI工具

### 前端
- **Vue 3** - 前端框架
- **Element Plus** - UI组件库
- **TypeScript** - 类型支持
- **Vite** - 构建工具
- **ECharts** - 数据可视化

## 快速开始

### 1. 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 6.0+

### 2. 配置文件

复制并修改配置文件：

```bash
cp configs/config.yaml configs/config.local.yaml
```

编辑 `configs/config.local.yaml`：

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: "your-password"
  dbname: go_market_email

redis:
  host: localhost
  port: 6379
  password: ""

smtp:
  host: smtp.partner.outlook.cn
  port: 587
  username: "your-email@outlook.com"
  password: "your-password"

ai:
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4"

auth:
  token: "your-secret-token"
```

### 3. 安装依赖

```bash
# 后端依赖
go mod tidy

# 前端依赖
cd web && npm install
```

### 4. 数据库初始化

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE go_market_email CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 5. 启动服务

```bash
# 启动后端服务
go run cmd/server/main.go -c configs/config.local.yaml

# 启动前端开发服务器
cd web && npm run dev

# 启动邮件发送工作进程
go run cmd/cli/main.go worker -c configs/config.local.yaml
```

## Docker部署

### 1. 构建镜像

```bash
docker build -t go-market-email .
```

### 2. 使用docker-compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## API文档

### 认证

所有API请求需要在Header中包含认证Token：

```
Authorization: Bearer your-secret-token
```

### 主要接口

#### 模板管理

```bash
# 创建模板
POST /api/v1/templates
{
  "name": "欢迎邮件",
  "subject": "欢迎 {{name}} 加入我们！",
  "content": "亲爱的 {{name}}，欢迎来到 {{company}}！",
  "project_id": 1
}

# 获取模板
GET /api/v1/templates/{id}

# 模板列表
GET /api/v1/templates?project_id=1&page=1&page_size=10
```

#### 邮件发送

```bash
# 测试发送
POST /api/v1/emails/test
{
  "template_id": 1,
  "email": "test@example.com",
  "data": {
    "name": "张三",
    "company": "ABC公司"
  }
}

# 创建批量任务
POST /api/v1/emails/tasks
{
  "name": "营销活动1",
  "template_id": 1,
  "data_source": "excel",
  "ai_prompt": "根据用户信息 {{name}} 和 {{city}} 生成个性化推荐"
}
```

## 环境变量

系统支持通过环境变量覆盖配置文件：

```bash
export GME_DATABASE_PASSWORD="your-db-password"
export GME_REDIS_PASSWORD="your-redis-password"
export GME_SMTP_PASSWORD="your-smtp-password"
export GME_AI_OPENAI_API_KEY="your-openai-key"
export GME_AUTH_TOKEN="your-auth-token"
```

## 监控和日志

### 日志配置

```yaml
log:
  level: info          # debug, info, warn, error
  retention_days: 7    # 日志保留天数
  file_path: "./logs/app.log"
```

### 监控指标

- 邮件模板数量
- 待发送邮件数量
- 发送成功/失败统计
- 平均发送耗时
- 预计完成时间

## 开发指南

### 项目结构

```
go_market_email/
├── cmd/                 # 命令行工具
│   ├── server/         # Web服务器
│   └── cli/            # CLI工具
├── internal/           # 内部包
│   ├── models/         # 数据模型
│   ├── services/       # 业务逻辑
│   ├── handlers/       # HTTP处理器
│   ├── middleware/     # 中间件
│   └── utils/          # 工具函数
├── web/                # 前端代码
│   ├── src/           # 源代码
│   └── dist/          # 构建产物
├── configs/           # 配置文件
├── docker/            # Docker相关
└── docs/              # 文档
```

### 添加新功能

1. 在 `internal/models/` 中定义数据模型
2. 在 `internal/services/` 中实现业务逻辑
3. 在 `internal/handlers/` 中添加HTTP处理器
4. 在前端 `web/src/` 中添加页面和组件

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查MySQL服务是否启动
   - 验证数据库配置信息
   - 确认数据库已创建

2. **Redis连接失败**
   - 检查Redis服务状态
   - 验证Redis配置

3. **邮件发送失败**
   - 检查SMTP配置
   - 验证邮箱密码（可能需要应用专用密码）
   - 确认网络连接

4. **前端无法访问API**
   - 检查后端服务是否启动
   - 验证CORS配置
   - 确认认证Token

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！