# Vercel 部署说明（详细）

本文档说明如何把本仓库的前端部署到 Vercel，并通过 GitHub Actions 自动触发部署。

## 在 Vercel 上创建项目

1. 登录 Vercel 并点击 "New Project" → 导入你的仓库。
2. 在 Import 页面设置 Root Directory 为 `frontend`（若仓库同时包含后端）。
3. 在 Build & Output 设置中确认：
   - Install Command: `pnpm install`
   - Build Command: `pnpm build`

## 设置 Environment Variables

在 Vercel 项目 Settings → Environment Variables 添加：

- `NEXT_PUBLIC_API_URL` (Production/Preview/Development) 指向后端 API，例：`https://api.example.com`
- `NEXT_PUBLIC_WS_URL` 指向后端 WebSocket，例：`wss://api.example.com/ws`
- `NEXT_PUBLIC_FRONTEND_URL` = `https://app.example.com`
- `ALCHEMY_API_KEY`（若 SSR 需要）

注意：把敏感密钥只放在 Server-side（不以 `NEXT_PUBLIC_` 开头）的变量，或使用 Vercel Secrets。

## 生成 Vercel Token（用于 CI）

1. 在 Vercel 控制台点击右上角用户头像 → Settings → Tokens → Create Token。
2. 复制生成的 token，去 GitHub 仓库 Settings → Secrets → Actions，添加 `VERCEL_TOKEN`。
3. 可选：添加 `VERCEL_PROJECT_ID` 与 `VERCEL_ORG_ID` 为 GitHub Secrets（有些部署方式需要）。

## 在 GitHub Actions 中部署（概述）

- 工作流会在 `main` 分支推送时构建 `frontend` 并使用 `VERCEL_TOKEN` 执行 `npx vercel --prod` 部署。
- 确保 Vercel 项目的 Environment Variables 已预先设置，或在工作流中注入（不建议把 secrets 写入仓库）。

## CORS 与 Preview 域名

- Vercel 的 Preview 部署会生成临时域名，若需要在 Preview 正常访问后端，请把这些 Preview 域名加入后端的 `CORS_ORIGINS`，或在后端短期允许 `https://vercel.app` 的请求。

## WebSocket 注意

- Vercel 的 Rewrite/Proxy 不适用于 WebSocket。前端应直接连接 `wss://` 后端域名，后端需部署到支持 WebSocket 的主机（ALB/Traefik/NGINX 并配置 sticky session 或使用 Redis 广播）。

## 回滚

- Vercel 在项目的 Deployments 页面可以直接回滚到任一历史部署。

## 常见问题

- 部署失败：在 Vercel 的 Deploy Logs 查看具体错误，常见为依赖安装、构建脚本或缺失环境变量。
- 需要在构建时访问私有 API：考虑在 Vercel 中配置 Server Environment Variables 并在 SSR 中安全使用。
