# 使用 Docker Buildx（BuildKit）

本文档说明如何在本地安装并启用 Docker Buildx（BuildKit），以及在本项目中如何使用。

## 为什么要用 Buildx

- Buildx（基于 BuildKit）支持并行构建、更高效缓存、多平台镜像与更灵活的构建器。
- Docker 官方已弃用 legacy builder，建议迁移到 buildx。

## 本地安装（macOS / Linux）

大多数新版 Docker Desktop 已自带 `buildx`。若 `docker buildx version` 提示未找到，可手动安装：

```bash
mkdir -p ~/.docker/cli-plugins
curl -L "https://github.com/docker/buildx/releases/latest/download/docker-buildx-$(uname -s)-$(uname -m)" \
  -o ~/.docker/cli-plugins/docker-buildx
chmod +x ~/.docker/cli-plugins/docker-buildx
# 验证
docker buildx version
```

## 启用并创建 builder

```bash
# 创建并使用名为 cf-builder 的 builder
docker buildx create --name cf-builder --use
# 确认 builder 状态
docker buildx ls
```

## 在本项目中使用

- 我已在 `Makefile` 中添加 `ensure-buildx` 目标，会在需要时自动创建 builder。
- 使用 `make docker-build` 会执行 `docker buildx build --push ...`，以 `latest` 和当前 git sha 两个 tag 推送到 registry。

注册表示例说明：

在中国大陆常见的 registry 示例（替换为你的实际 registry）：

- 阿里云 ACR: `registry.cn-<region>.aliyuncs.com/<namespace>`，例如 `registry.cn-hangzhou.aliyuncs.com/myteam`。

示例登录与构建流程（本地）：

```bash
# 登录阿里云镜像仓库
docker login --username=<your-username> registry.cn-hangzhou.aliyuncs.com

# 本地构建并加载到 docker（不会推送）
make docker-build

# 推送（确保 DOCKER_REGISTRY/凭据已设置）
export DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com/myteam
export DOCKER_REGISTRY_USER=<your-username>
export DOCKER_REGISTRY_PASSWORD=<your-password-or-token>
make docker-push
```

在 CI 中，请把 `DOCKER_REGISTRY`、`DOCKER_REGISTRY_USER` 与 `DOCKER_REGISTRY_PASSWORD` 作为 GitHub Secrets 设置，并在 workflow 中使用 `docker/login-action` 进行登录后再用 `docker/build-push-action` 进行构建/推送。

本地测试（无需推送）可用：

```bash
# 临时启用 BuildKit
export DOCKER_BUILDKIT=1
# 构建并加载到本地 docker（注意：--load 需本地 builder 支持）
docker buildx build --platform linux/amd64 -t chainfeed-local:latest -f Dockerfile --load .
```

## CI（GitHub Actions）

- `.github/workflows/backend.yml` 已包含 `docker/setup-buildx-action` 与 `docker/build-push-action`，可以在 Actions 中使用 buildx 无需额外手动安装。

## 常见问题

- 错误：`failed to solve: rpc error: code = Unknown desc = no available session`: 可能是 buildx 需要创建 builder 或 Docker daemon 未正确运行，尝试 `docker buildx create --use`。
- 在 macOS 使用 Docker Desktop 通常无需额外安装，确保 Docker Desktop 已升级到最新版本。

