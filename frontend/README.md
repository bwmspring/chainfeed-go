## 部署（Vercel）

推荐使用 Vercel 部署 `frontend` 目录的 Next.js 应用。下面为最少可行的步骤与环境变量示例。

1. 连接仓库
   - 在 Vercel 控制台创建新项目，选择你的仓库。Import 时把 Root Directory 设置为 `frontend`（如果仓库包含后端）。

2. 必填环境变量（在 Vercel → Project → Settings → Environment Variables 添加）
   - `NEXT_PUBLIC_API_URL` = https://api.example.com
   - `NEXT_PUBLIC_WS_URL` = wss://api.example.com/ws
   - `NEXT_PUBLIC_FRONTEND_URL` = https://app.example.com

3. 可选/构建时环境变量（Server-side，勿以 NEXT_PUBLIC_ 开头）
   - `API_URL` = https://api.example.com
   - `ALCHEMY_API_KEY` = <alchemy_key>（如果在 SSR 或构建时使用）

4. 构建设置
   - Install Command: `pnpm install`
   - Build Command: `pnpm build`
   - Output Directory: 留空（Next.js）

5. 部署与本地验证
```bash
cd frontend
pnpm install
pnpm build
pnpm start # 本地生产预览
# Vercel 本地预览（需安装 vercel CLI）
npx vercel dev
```

6. Preview 环境
   - 在 Vercel 中对 `Preview` 分支也设置相应的环境变量，或允许 Preview 域名在后端 `CORS_ORIGINS` 中白名单。

注意
- 不要把 `JWT_SECRET`、`WEBHOOK_SECRET` 放到前端环境变量。
- 若前端需直接访问第三方需要密钥的 API（如 Alchemy），优先使用后端代理以保护密钥。
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
