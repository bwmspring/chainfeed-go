# ChainFeed Go 项目结构

## 目录结构

```
chainfeed-go/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go        # 主程序入口
├── internal/              # 内部包（不对外暴露）
│   ├── config/           # 配置管理
│   │   └── config.go
│   ├── database/         # 数据库连接
│   │   ├── postgres.go
│   │   └── redis.go
│   └── server/           # HTTP 服务器
│       └── server.go
├── pkg/                   # 公共包（可对外暴露）
│   └── logger/           # 日志模块
│       └── logger.go
├── migrations/           # 数据库迁移文件
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── config/              # 配置文件
│   ├── config.yaml
│   └── config.test.yaml
├── logs/                # 日志文件目录
├── bin/                 # 编译后的二进制文件
├── docker-compose.yml   # Docker 服务编排
├── Makefile            # 构建脚本
├── go.mod              # Go 模块定义
├── .gitignore          # Git 忽略文件
├── .golangci.yml       # 代码检查配置
└── .env.example        # 环境变量示例
```

## 核心特性

### 1. Web3 认证系统
- 钱包地址作为用户标识
- 基于签名的无密码认证
- Nonce 防重放攻击
- JWT Token 会话管理

### 2. 配置管理
- 使用 Viper 支持多种配置格式 (YAML, JSON, ENV)
- 支持环境变量覆盖
- 分层配置结构，便于管理

### 2. 日志系统
- 基于 Zap 高性能日志库
- 支持结构化日志 (JSON) 和控制台输出
- 可配置日志级别和输出路径
- 生产环境友好的日志格式

### 3. 数据库连接
- PostgreSQL 连接池管理
- Redis 连接池管理
- 连接健康检查
- 优雅的错误处理

### 4. HTTP 服务器
- 基于 Gin 框架
- 内置健康检查端点
- 优雅关闭支持
- 中间件支持

### 5. 开发工具
- Makefile 自动化构建
- Docker Compose 本地开发环境
- golangci-lint 代码质量检查
- 数据库迁移支持

## 使用方法

### 1. 启动开发环境
```bash
# 启动数据库服务
make docker-up

# 运行数据库迁移
make migrate-up

# 启动应用
make run
```

### 2. 构建和测试
```bash
# 构建应用
make build

# 运行测试
make test

# 代码检查
make lint
```

### 3. 健康检查
```bash
curl http://localhost:8080/health
```

## 生产级别特性

1. **错误处理**: 统一的错误处理和日志记录
2. **配置管理**: 支持多环境配置
3. **连接池**: 数据库连接池优化
4. **优雅关闭**: 支持 SIGTERM/SIGINT 信号处理
5. **健康检查**: 内置健康检查端点
6. **日志轮转**: 支持日志文件输出
7. **代码质量**: golangci-lint 静态分析
8. **容器化**: Docker 支持

## 下一步

Phase 1.1 基础设施搭建已完成，包括：
- ✅ 项目脚手架初始化
- ✅ 数据库表结构设计与迁移脚本
- ✅ Redis 连接与配置
- ✅ 基础 API 路由框架
- ✅ 配置管理系统
- ✅ 日志系统
- ✅ 构建和开发工具

可以开始 Phase 1.2: 数据索引层开发。
