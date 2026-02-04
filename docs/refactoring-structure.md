# 重构后的项目结构

## 📁 目录结构

```
chainfeed-go/
├── cmd/server/
│   └── main.go                 # 简洁的程序入口 (15 行)
├── internal/
│   ├── app/
│   │   └── app.go             # 应用初始化和生命周期管理
│   ├── routes/
│   │   ├── api.go             # API 路由 (用户、地址、Feed)
│   │   └── webhook.go         # Webhook 和监控路由
│   ├── server/
│   │   └── server.go          # HTTP 服务器 (简化版)
│   ├── config/
│   ├── database/
│   ├── models/
│   ├── repository/
│   ├── parser/
│   └── webhook/
└── ...
```

## 🔄 重构优势

### 1. 清晰的关注点分离
```go
// main.go - 只负责启动
func main() {
    app := app.New(configPath)
    app.Run()
}

// app.go - 负责初始化和生命周期
func (a *App) Run() error {
    // 启动服务器
    // 处理信号
    // 优雅关闭
}

// routes/*.go - 负责路由定义
func (r *APIRoutes) RegisterRoutes(router *gin.RouterGroup) {
    // 注册具体路由
}
```

### 2. 模块化路由管理
```go
// API 路由模块
/api/v1/users/auth          # Web3 认证
/api/v1/users/profile       # 用户资料
/api/v1/addresses           # 监控地址管理
/api/v1/feed                # 交易 Feed

// Webhook 路由模块  
/webhooks/alchemy           # Alchemy Webhook

// 监控路由模块
/monitoring/stats           # 性能统计
```

### 3. 易于扩展的架构
```go
// 添加新的路由模块
type UserRoutes struct { /* ... */ }
func (r *UserRoutes) RegisterRoutes(router *gin.RouterGroup) { /* ... */ }

// 在 server.go 中注册
userRoutes := routes.NewUserRoutes(...)
userRoutes.RegisterRoutes(s.router.Group(""))
```

## 🚀 API 端点规划

### 用户认证 (Phase 1.3)
```
POST /api/v1/users/auth
GET  /api/v1/users/profile
```

### 地址管理 (Phase 1.3)
```
GET    /api/v1/addresses           # 获取监控地址列表
POST   /api/v1/addresses           # 添加监控地址
DELETE /api/v1/addresses/:id       # 删除监控地址
```

### 交易 Feed (Phase 1.4)
```
GET /api/v1/feed                   # 获取用户 Feed 流
GET /api/v1/feed/transactions/:hash # 获取交易详情
```

### 系统监控
```
GET /health                        # 健康检查
GET /monitoring/stats              # 性能统计
```

## 📊 代码行数对比

### 重构前
```
cmd/server/main.go: 80+ 行 (包含所有初始化逻辑)
internal/server/server.go: 120+ 行 (包含所有路由)
```

### 重构后
```
cmd/server/main.go: 15 行 (只负责启动)
internal/app/app.go: 80 行 (初始化逻辑)
internal/server/server.go: 60 行 (核心服务器逻辑)
internal/routes/api.go: 60 行 (API 路由)
internal/routes/webhook.go: 40 行 (Webhook 路由)
```

## 🎯 优势总结

1. **可维护性**: 每个文件职责单一，易于理解和修改
2. **可扩展性**: 新增路由模块不影响现有代码
3. **可测试性**: 每个模块可以独立测试
4. **团队协作**: 不同开发者可以并行开发不同模块
5. **代码复用**: 路由模块可以在不同项目中复用

这种结构为后续的 Phase 1.3 (用户管理) 和 Phase 1.4 (Feed 流) 开发奠定了良好的基础。
