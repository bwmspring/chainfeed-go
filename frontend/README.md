# ChainFeed Frontend

## 技术栈

- ✅ Next.js 14 (App Router)
- ✅ TypeScript
- ✅ TailwindCSS + shadcn/ui
- ✅ wagmi + viem (钱包连接)
- ✅ RainbowKit (钱包 UI)
- ✅ TanStack Query (数据获取)
- ✅ WebSocket (实时 Feed)

## 快速开始

### 1. 安装依赖
```bash
pnpm install
```

### 2. 配置环境变量
复制 `.env.local` 并填写：
- `NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID`: 从 [WalletConnect Cloud](https://cloud.walletconnect.com/) 获取

### 3. 启动开发服务器
```bash
pnpm dev
```

访问 http://localhost:3000

## 项目结构

```
frontend/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # 根布局（集成 Providers）
│   └── page.tsx           # 主页（Feed 流）
├── components/
│   ├── ui/                # shadcn/ui 组件
│   ├── providers.tsx      # Web3 + Query Providers
│   ├── header.tsx         # 顶部导航
│   └── feed-card.tsx      # Feed 卡片
├── hooks/
│   ├── use-websocket.ts   # WebSocket 连接
│   └── use-api.ts         # API 请求 hooks
└── lib/
    ├── wagmi.ts           # wagmi 配置
    └── utils.ts           # 工具函数
```

## 核心功能

### 钱包连接
使用 RainbowKit 提供的 `ConnectButton` 组件，支持多种钱包。

### 实时 Feed
通过 WebSocket 接收链上活动推送，自动更新 Feed 流。

### API 集成
使用 TanStack Query 管理服务端状态，自动缓存和重新验证。

## 开发指南

### 添加新组件
```bash
pnpm dlx shadcn@latest add [component-name]
```

### 类型安全
所有 API 响应和 WebSocket 消息都应定义 TypeScript 类型。

### 状态管理
- 服务端状态：TanStack Query
- 客户端状态：React useState/useReducer
- Web3 状态：wagmi hooks
