# ChainFeed (Coming Soon) ⛓️

像刷 Twitter 一样追踪链上活动 —— 专注于 Web3 链上数据的实时信息流平台。

> 概念演示 (Coming Soon) • 项目文档 (Coming Soon)

---

## 📖 项目愿景（Overview）

在 Web3 世界，链上数据虽然公开透明，但对普通开发者和投资者而言依然存在“信息围墙”。ChainFeed 旨在打破这一壁垒。

我们构建一个高性能的后端索引层，将复杂的链上交易（Logs / Events）解析为人类可读的社交化信息流。用户只需关注特定地址（如巨鲸、KOL 或机构），即可实时掌握其资产流转与交互动态。

## 核心价值

- **实时性**：通过 WebSocket 建立长连接，保证链上动态的秒级感知。
- **可读性**：使用 AI（GPT-4o 等）自动解析复杂交易逻辑，将十六进制数据转换为自然语言摘要。
- **专注性**：聚焦以太坊主网，深度解析链上活动与交互模式。

## ✨ 规划功能（Planned Features）

### 1.0 MVP

- 智能关注管理：地址标签、ENS 解析与批量监控。
- 实时 Feed 流：覆盖 ETH、ERC-20 转账与 NFT（ERC-721）交易。
- AI 交易摘要：为复杂合约交互生成简短精准的中文说明。
- 以太坊主网深度解析：专注单链，提供最精准的链上活动监控。

### 进阶功能

- 异常行为监测：大额异动告警与批量授权风险预警。
- 链上关系图谱：分析地址间资金往来与关联性。
- 数据导出 API：为开发者提供实时数据推送服务。

## 🏗️ 技术架构（Architecture Design）

本项目追求高并发处理能力与良好的可观测性，核心组件包括后台索引器、AI 解析引擎、以及实时推送层。

### 技术栈

**后端**
- Go 1.22 + Gin
- PostgreSQL + Redis
- Alchemy Webhooks + go-ethereum
- Dify + OpenAI GPT-4o

**前端**
- Next.js 14 (App Router) + TypeScript
- TailwindCSS + shadcn/ui
- wagmi + viem + RainbowKit (Web3)
- TanStack Query + WebSocket

### 数据流（示意）

```mermaid
graph LR
    A[Blockchain Networks] --> B[Data Streams Layer]
    B -->|Webhook| C[Go Backend Workers]
    C --> D[(Postgres / Redis)]
    C --> E[AI Analysis Engine]
    D --> F[WebSocket Service]
    F --> G[Frontend Feed UI]
```

## 📊 技术难点（Engineering Challenges）

- **高并发下的数据一致性**：在高频 Webhook 推送场景，如何通过 Redis 队列和幂等策略保证入库顺序与准确性。
- **智能合约解析**：深度解析以太坊主网复杂合约交互，提供精准的交易语义理解。
- **低延迟推送**：优化 WebSocket 服务端的内存与连接管理，支撑大量实时连接。


## 🚀 快速开始

### 环境要求

**后端**
- Go 1.22+
- PostgreSQL 14+
- Redis 7+
- Docker & Docker Compose (可选)

**前端**
- Node.js 18+
- pnpm (推荐)


### 手动启动

#### 后端

```bash
# 1. 启动数据库
docker-compose up -d

# 2. 运行数据库迁移
make migrate-up

# 3. 启动后端服务
make run
```

#### 前端

```bash
cd frontend

# 1. 安装依赖
pnpm install

# 2. 配置环境变量
# 复制 .env.local 并填写 NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID
# 从 https://cloud.walletconnect.com/ 获取

# 3. 启动开发服务器
pnpm dev
```

访问 http://localhost:3000

## 📚 API 文档

项目集成了 Swagger API 文档，提供完整的接口说明和在线测试功能。

### ⚠️ 当前状态

由于网络问题，Swagger 依赖暂未安装，文档 UI 暂时不可用。

**临时方案：**
- 查看 API 注释：所有接口都有完整的 Swagger 注释
- 使用文档：`docs/phase-1.3-quickstart.md` 包含所有 API 使用示例
- 使用 curl/Postman 测试接口

### 启用 Swagger（网络恢复后）

```bash
# 1. 安装 Swagger 工具
make swagger-install

# 2. 取消 internal/routes/api.go 中的注释

# 3. 生成文档
make swagger

# 4. 启动服务
make run

# 5. 访问文档
open http://localhost:8080/swagger/index.html
```

### 主要接口

- **认证**：`POST /api/v1/auth/nonce`、`POST /api/v1/auth/verify`
- **用户**：`GET /api/v1/profile`
- **监控地址**：`GET/POST/DELETE /api/v1/addresses`

## ✉️ 联系方式

如果你对本项目感兴趣或希望参与协作，欢迎联系：

- Email: bwm029@gmail.com

---

ChainFeed — 洞察链上脉搏，让复杂数据触手可及。
