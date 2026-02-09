# Secrets 管理指南（快速上手）

本文档目标：列出本项目所需的 Secret、给出在 GitHub 与 Vercel 中的配置步骤，以及在自托管主机上使用 Docker Secrets 的示例。

一、必须管理的 Secret（最小集合）

- `DB_PASSWORD`：Postgres 密码（生产数据库访问凭据）
- `JWT_SECRET`：后端 JWT 签名密钥
- `WEBHOOK_SECRET`：外部 webhook 验证密钥
- `REDIS_PASSWORD`：若 Redis 有密码则需配置
- `ALCHEMY_API_KEY`：Alchemy RPC / API Key
- `DOCKER_REGISTRY_USER` / `DOCKER_REGISTRY_PASSWORD`：推镜像用的仓库凭据（例如 ghcr / Docker Hub）
- `SSH_PRIVATE_KEY`：CI 用于 SSH 到部署主机的私钥（仅在 CI Secrets 中使用）
- `VERCEL_TOKEN`：Vercel 自动部署 token（用于 GitHub Actions 部署）

二、在 GitHub 仓库设置 Secrets（推荐用于 CI）

1. 打开仓库 → `Settings` → `Secrets and variables` → `Actions` → `New repository secret`。
2. 逐个添加上面的秘密名称及其值（名称区分大小写）。
3. 在 Actions Workflow 中引用：

```yaml
env:
  VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
  DOCKER_REGISTRY: ${{ secrets.DOCKER_REGISTRY }}
```

示例：使用 GitHub CLI 添加（本地操作）
```bash
gh secret set VERCEL_TOKEN --body "$VERCEL_TOKEN"
gh secret set DOCKER_REGISTRY_PASSWORD --body "$DOCKER_REGISTRY_PASSWORD"
```

三、在 Vercel 添加环境变量（用于前端与构建时）

1. 在 Vercel 控制台打开项目 → `Settings` → `Environment Variables`。
2. 添加变量：
   - 浏览器可见的变量需以 `NEXT_PUBLIC_` 前缀，例如 `NEXT_PUBLIC_API_URL`、`NEXT_PUBLIC_WS_URL`。
   - 构建/服务器端变量（不公开）直接添加例如 `ALCHEMY_API_KEY`。
3. 为 `Production` / `Preview` / `Development` 分别设置合适的值。

注意：不要在前端添加 `JWT_SECRET`、`WEBHOOK_SECRET` 等后端专属的敏感项。

四、在自托管生产主机用 Docker Secrets（适用于 Swarm 或使用 compose v3.5+）

示例：
```bash
# 创建 secret（在管理节点上）
echo "my_db_password" | docker secret create db_password -
echo "my_jwt_secret" | docker secret create jwt_secret -

# 在 docker-compose.prod.yml 中使用（已在仓库提供示例）
# services.backend.secrets: - db_password
```

说明：Docker Secrets 会以文件形式在容器内 `/run/secrets/<name>` 出现，应用需读取该文件（或在 compose 中由环境变量读取）。

五、最佳实践（必须遵守）

- 不要在仓库中提交任何真实的 `.env` 或 secret。保留并提交 `.env.example` 作为文档。
- 在 CI 中把 secret 限定为仅在必要的工作流/分支可用（例如 Production 环境变量仅在 `main` 的 workflow 使用）。
- 使用短期/最小权限的凭据（例如 CI 用的 registry token 只允许 push 权限）。
- 定期轮换关键密钥（JWT、DB 密码），并记录变更与回滚步骤。

六、逐步落地建议（快速可执行方案）

1. 立即：在 GitHub Secrets 中添加 `DOCKER_REGISTRY_*`、`VERCEL_TOKEN`、`SSH_PRIVATE_KEY`、`ALCHEMY_API_KEY`、`JWT_SECRET`、`WEBHOOK_SECRET`。
2. 在 Vercel 项目中添加 `NEXT_PUBLIC_API_URL`、`NEXT_PUBLIC_WS_URL`、`ALCHEMY_API_KEY`（若 SSR 需要）。
3. 在生产主机使用 Docker Secrets 管理运行时密码（`db_password`、`redis_password`、`jwt_secret`）。

七、示例：在 CI 中安全使用 Secret（关键点）

- 不打印 secret 值到日志（在 workflow 中避免 echo secrets）。
- 使用 `actions/checkout` 的 `persist-credentials: false`（按需），并仅在需要时把私钥写入文件执行 SSH，然后删除。

八、资源与延伸阅读

- GitHub Actions Secrets: https://docs.github.com/en/actions/security-guides/encrypted-secrets
- Vercel Environment Variables: https://vercel.com/docs/concepts/projects/environment-variables
- Docker Secrets: https://docs.docker.com/engine/swarm/secrets/
