.PHONY: help build run test clean swagger swagger-install deps fmt imports golines format

help:
	@echo "ChainFeed - 可用命令:"
	@echo "  make build       - 编译项目"
	@echo "  make run         - 运行服务"
	@echo "  make test        - 运行测试"
	@echo "  make clean       - 清理构建文件"
	@echo "  make deps        - 安装项目依赖"
	@echo "  make swagger     - 生成 Swagger 文档"
	@echo "  make dev         - 生成文档并运行服务"
	@echo "  make migrate     - 运行数据库迁移"
	@echo "  make migrate-down- 回滚数据库迁移"
	@echo "  make db-reset    - 重置数据库"
	@echo "  make fmt         - 格式化代码"
	@echo "  make imports     - 整理 import"
	@echo "  make golines     - 格式化长行"
	@echo "  make format      - 完整格式化 (fmt + imports + golines)"

build:
	@go build -o bin/chainfeed cmd/server/main.go

run:
	@go run cmd/server/main.go

dev: swagger run

test:
	@go test -v ./...

clean:
	@rm -rf bin/ docs/swagger/

deps:
	@export GOPROXY=https://goproxy.cn,direct && \
	go mod download && go mod tidy

swagger-install:
	@export GOPROXY=https://goproxy.cn,direct && \
	go install github.com/swaggo/swag/cmd/swag@latest && \
	go get -u github.com/swaggo/gin-swagger github.com/swaggo/files && \
	go mod tidy

swagger:
	@swag init -g cmd/server/main.go -o docs/swagger


migrate:
	@echo "运行数据库迁移..."
	@PGPASSWORD=chainfeed psql -h localhost -p 5432 -U chainfeed -d chainfeed < migrations/000001_init_schema.up.sql
	@echo "✅ 迁移完成"

migrate-down:
	@echo "回滚数据库迁移..."
	@PGPASSWORD=chainfeed psql -h localhost -p 5432 -U chainfeed -d chainfeed < migrations/000001_init_schema.down.sql
	@echo "✅ 回滚完成"

db-reset: migrate-down migrate

fmt:
	@go fmt ./...

imports:
	@if ! command -v goimports &> /dev/null; then \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@goimports -w .

golines:
	@if ! command -v golines &> /dev/null; then \
		go install github.com/segmentio/golines@latest; \
	fi
	@golines -w --max-len=120 --base-formatter=gofmt .

format: fmt imports golines

lint:
	@golangci-lint run

install: deps swagger-install

# Docker / Deploy variables (override as needed)
DOCKER_REGISTRY ?= registry.cn-hangzhou.aliyuncs.com/myteam
IMAGE_NAME ?= $(DOCKER_REGISTRY)/chainfeed
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)


.PHONY: docker-build docker-push deploy-remote


# docker-build: only build image locally (no push)
.PHONY: docker-build
docker-build:
	@echo "Building image $(IMAGE_NAME):$(IMAGE_TAG) (no push)"
	@echo "Using buildx to build and load image locally"
	@docker buildx build --platform linux/amd64 \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		--load -f Dockerfile .

# docker-push: push previously built image to registry
.PHONY: docker-push
docker-push:
	@echo "Pushing image $(IMAGE_NAME):$(IMAGE_TAG) to $(DOCKER_REGISTRY)"
	@if [ -z "$(DOCKER_REGISTRY_USER)" ] || [ -z "$(DOCKER_REGISTRY_PASSWORD)" ]; then \
		echo "Please set DOCKER_REGISTRY_USER and DOCKER_REGISTRY_PASSWORD environment variables."; exit 1; \
	fi
	@echo "Logging into registry $(DOCKER_REGISTRY)"
	@echo "$(DOCKER_REGISTRY_PASSWORD)" | docker login $(DOCKER_REGISTRY) -u "$(DOCKER_REGISTRY_USER)" --password-stdin
	@docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@docker push $(IMAGE_NAME):latest

deploy-remote: docker-push
	@echo "Deploying to remote host via SSH"
	@if [ -z "$(SSH_HOST)" ] || [ -z "$(SSH_USER)" ]; then \
		echo "Please set SSH_HOST and SSH_USER environment variables for remote deploy."; exit 1; \
	fi
	@echo "Running remote deploy commands on $(SSH_USER)@$(SSH_HOST)"
	@ssh -o StrictHostKeyChecking=no $(SSH_USER)@$(SSH_HOST) '\\
		docker pull $(IMAGE_NAME):$(IMAGE_TAG) && \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) chainfeed-backend:latest && \
		docker-compose -f $(REMOTE_COMPOSE_PATH) up -d backend \
	'

