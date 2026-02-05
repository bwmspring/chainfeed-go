.PHONY: help build run test clean swagger swagger-install deps fmt imports golines format

help:
	@echo "ChainFeed - 可用命令:"
	@echo "  make build    - 编译项目"
	@echo "  make run      - 运行服务"
	@echo "  make test     - 运行测试"
	@echo "  make clean    - 清理构建文件"
	@echo "  make deps     - 安装项目依赖"
	@echo "  make swagger  - 生成 Swagger 文档"
	@echo "  make dev      - 生成文档并运行服务"
	@echo "  make fmt      - 格式化代码"
	@echo "  make imports  - 整理 import"
	@echo "  make golines  - 格式化长行"
	@echo "  make format   - 完整格式化 (fmt + imports + golines)"

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
	@psql -U postgres -d chainfeed < migrations/001_create_transactions_table.sql
	@psql -U postgres -d chainfeed < migrations/002_add_indexes.sql
	@psql -U postgres -d chainfeed < migrations/003_create_user_tables.sql

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
