# 后端 CI/CD 与 远端部署说明

此文档说明如何使用仓库中提供的 GitHub Actions（`.github/workflows/backend.yml`）及 `Makefile` 目标来自动构建、推送后端镜像并在远端主机更新服务。

必要的 GitHub Secrets（在仓库 Settings → Secrets and variables → Actions 添加）：

- `DOCKER_REGISTRY`：镜像仓库地址，例如 `ghcr.io/your-org`（也可在 workflow 中使用完整地址）。
- `DOCKER_REGISTRY_USER` 与 `DOCKER_REGISTRY_PASSWORD`：用于登录镜像仓库的凭据或 token。
- 可选：`SSH_HOST`、`SSH_USER`、`SSH_PRIVATE_KEY`：用于在镜像推送后通过 SSH 在远端主机上拉取并更新服务。

使用示例（本地）：

```bash
# 构建并推送镜像（本地）
export DOCKER_REGISTRY=ghcr.io/your-org
export DOCKER_REGISTRY_USER=youruser
export DOCKER_REGISTRY_PASSWORD=yourtoken
make docker-push

# 远端部署（需在远端主机上准备 docker-compose.prod.yml）
export SSH_HOST=your.server.example.com
export SSH_USER=ubuntu
export REMOTE_COMPOSE_PATH=/home/ubuntu/chainfeed/docker-compose.prod.yml
make deploy-remote
```

CI 行为（示例工作流）

- 当 `main` 分支有变更（限制为后端相关路径）时，Workflow 会：
  1. 检出代码并设置 Go 环境
  2. 使用 `docker/build-push-action` 通过 Buildx 构建并推送多平台镜像到 `DOCKER_REGISTRY`
  3. （可选）若设置了 SSH Secrets，则通过 SSH 在远端主机执行拉取镜像并用 `docker-compose` 更新后端服务

安全建议：

- 把所有敏感凭据放到 GitHub Secrets，不要在仓库或构建日志中输出它们。
- 远端主机应使用受限用户与私钥并开启防火墙，仅允许 CI 的 IP 或管理 IP 访问（若可能）。

国内镜像仓库示例 — 阿里云 ACR

若你在中国大陆部署，推荐使用阿里云容器镜像服务（ACR）作为镜像仓库，网络访问速度快且稳定。

基本流程：

1. 在阿里云控制台 → 容器镜像服务（Container Registry）中新建命名空间（Namespace）和仓库（Repository）。
2. 获取推送凭据：可以使用主账号 AccessKey（不推荐用于 CI），更安全的做法是创建一个 RAM 子账号并给其最小权限，使用该子账号的 AccessKey ID/Secret 作为登录凭据；也可以在控制台创建镜像仓库登录密码（token）。
3. 登录并推送（本地示例）：

```bash
# 假设你的命名空间为 `myteam`，Region 为 `cn-hangzhou`，仓库地址样例：
DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/myteam
docker login --username=<your-username> registry.cn-hangzhou.aliyuncs.com
# 输入密码或 token
docker tag chainfeed:latest ${DOCKER_REGISTRY}/chainfeed:latest
docker push ${DOCKER_REGISTRY}/chainfeed:latest
```

在本项目中（使用 Makefile）：

```bash
export DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/myteam
export DOCKER_REGISTRY_USER=<your-username>
export DOCKER_REGISTRY_PASSWORD=<your-token-or-password>
make docker-push
```

在 GitHub Actions 中设置 Secrets（示例）：

- `DOCKER_REGISTRY` = registry.cn-hangzhou.aliyuncs.com/myteam
- `DOCKER_REGISTRY_USER` = <your-ram-username>
- `DOCKER_REGISTRY_PASSWORD` = <the-access-key-secret-or-token>

可选：使用阿里云 CLI 创建仓库（如果你偏向命令行自动化）：

```bash
aliyun cr CreateRepository --RepositoryName chainfeed --RepoNamespace myteam --RegionId cn-hangzhou
```

注意事项：

- 请为 CI 使用的账号分配最小权限（仅允许 push/pull 对应仓库）。
- 若项目对外公开或多人协作，建议在控制台开启镜像扫描与镜像加速策略。
