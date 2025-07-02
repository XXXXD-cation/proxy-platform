# 高性能代理与隧道平台 - Makefile
# Author: ccnochch
# Description: 提供便民的开发和部署命令

.PHONY: help dev-setup dev-start dev-stop dev-clean build test lint format migrate-up migrate-down logs

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
DOCKER_COMPOSE := docker-compose
GO_CMD := go
MIGRATE_CMD := migrate
PROJECT_NAME := proxy-platform

# 帮助信息
help: ## 显示帮助信息
	@echo "高性能代理与隧道平台 - 开发命令"
	@echo "=================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ====================
# 开发环境管理
# ====================

dev-setup: ## 快速搭建开发环境
	@echo "🚀 正在搭建开发环境..."
	@echo "1. 检查Docker和Docker Compose..."
	@which docker > /dev/null || (echo "❌ Docker未安装" && exit 1)
	@which docker-compose > /dev/null || (echo "❌ Docker Compose未安装" && exit 1)
	@echo "2. 创建必要的目录..."
	@mkdir -p logs configs/{gateway,proxy-pool,admin-api,free-crawler}
	@mkdir -p deployments/{nginx/conf.d,redis,prometheus,grafana/{provisioning,dashboards}}
	@echo "3. 生成配置文件..."
	@$(MAKE) generate-configs
	@echo "4. 下载Go依赖..."
	@$(GO_CMD) mod download
	@echo "✅ 开发环境搭建完成！"
	@echo ""
	@echo "🎯 下一步操作："
	@echo "   make dev-start    # 启动开发环境"
	@echo "   make logs         # 查看服务日志"

dev-start: ## 启动开发环境
	@echo "🎬 启动开发环境..."
	@$(DOCKER_COMPOSE) up -d mysql redis
	@echo "⏳ 等待数据库启动..."
	@sleep 15
	@$(MAKE) migrate-up
	@$(DOCKER_COMPOSE) up -d
	@echo "✅ 开发环境启动完成！"
	@echo ""
	@echo "🌐 服务访问地址："
	@echo "   前端应用:      http://localhost:5173"
	@echo "   API网关:       http://localhost:8080"
	@echo "   管理后台API:   http://localhost:8082"
	@echo "   代理池服务:    http://localhost:8081"
	@echo "   免费爬虫:      http://localhost:8083"
	@echo "   监控面板:      http://localhost:3000 (admin/admin123)"
	@echo "   Prometheus:    http://localhost:9090"

dev-stop: ## 停止开发环境
	@echo "🛑 停止开发环境..."
	@$(DOCKER_COMPOSE) down
	@echo "✅ 开发环境已停止"

dev-restart: ## 重启开发环境
	@echo "🔄 重启开发环境..."
	@$(MAKE) dev-stop
	@$(MAKE) dev-start

dev-clean: ## 清理开发环境（包括数据）
	@echo "🧹 清理开发环境..."
	@echo "⚠️  这将删除所有容器、网络和数据卷！"
	@read -p "确认继续？(y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker system prune -f
	@echo "✅ 开发环境清理完成"

# ====================
# 代码构建和测试
# ====================

build: ## 构建所有服务
	@echo "🔨 构建所有服务..."
	@$(DOCKER_COMPOSE) build
	@echo "✅ 构建完成"

build-service: ## 构建指定服务 (使用: make build-service SERVICE=gateway)
	@echo "🔨 构建服务: $(SERVICE)"
	@$(DOCKER_COMPOSE) build $(SERVICE)

test: ## 运行单元测试
	@echo "🧪 运行单元测试..."
	@$(GO_CMD) test -v -race -coverprofile=coverage.out ./...
	@$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ 测试完成，覆盖率报告: coverage.html"

test-integration: ## 运行集成测试
	@echo "🧪 运行集成测试..."
	@$(GO_CMD) test -v -tags=integration ./tests/integration/...

lint: ## 代码静态检查
	@echo "🔍 代码静态检查..."
	@golangci-lint run ./...
	@echo "✅ 代码检查完成"

format: ## 代码格式化
	@echo "✨ 代码格式化..."
	@$(GO_CMD) fmt ./...
	@goimports -w .
	@echo "✅ 代码格式化完成"

# ====================
# 数据库迁移
# ====================

migrate-up: ## 执行数据库迁移
	@echo "📊 执行数据库迁移..."
	@sleep 5  # 等待数据库就绪
	@$(GO_CMD) run migrations/migrate.go up
	@echo "✅ 数据库迁移完成"

migrate-down: ## 回滚数据库迁移
	@echo "📊 回滚数据库迁移..."
	@$(GO_CMD) run migrations/migrate.go down
	@echo "✅ 数据库回滚完成"

migrate-create: ## 创建新的迁移文件 (使用: make migrate-create NAME=add_users_table)
	@echo "📝 创建迁移文件: $(NAME)"
	@$(GO_CMD) run migrations/migrate.go create $(NAME)

# ====================
# 日志和监控
# ====================

logs: ## 查看所有服务日志
	@$(DOCKER_COMPOSE) logs -f

logs-service: ## 查看指定服务日志 (使用: make logs-service SERVICE=gateway)
	@$(DOCKER_COMPOSE) logs -f $(SERVICE)

status: ## 查看服务状态
	@echo "📋 服务状态："
	@$(DOCKER_COMPOSE) ps

health: ## 健康检查
	@echo "🏥 健康检查..."
	@curl -f http://localhost:8080/health && echo "✅ 网关服务正常" || echo "❌ 网关服务异常"
	@curl -f http://localhost:8081/health && echo "✅ 代理池服务正常" || echo "❌ 代理池服务异常"
	@curl -f http://localhost:8082/health && echo "✅ 管理API服务正常" || echo "❌ 管理API服务异常"
	@curl -f http://localhost:8083/health && echo "✅ 免费爬虫服务正常" || echo "❌ 免费爬虫服务异常"

# ====================
# 工具函数
# ====================

generate-configs: ## 生成配置文件
	@echo "📝 生成配置文件..."
	@mkdir -p deployments/redis deployments/prometheus deployments/grafana/provisioning deployments/grafana/dashboards deployments/nginx/conf.d
	@echo "# Redis配置" > deployments/redis/redis.conf
	@echo "bind 0.0.0.0" >> deployments/redis/redis.conf
	@echo "port 6379" >> deployments/redis/redis.conf
	@echo "save 900 1" >> deployments/redis/redis.conf
	@echo "save 300 10" >> deployments/redis/redis.conf
	@echo "save 60 10000" >> deployments/redis/redis.conf
	@echo "# 基础配置文件已生成"

clean-logs: ## 清理日志文件
	@echo "🧹 清理日志文件..."
	@rm -rf logs/*
	@echo "✅ 日志清理完成"

# ====================
# 部署相关
# ====================

deploy-staging: ## 部署到测试环境
	@echo "🚀 部署到测试环境..."
	@# TODO: 实现测试环境部署逻辑

deploy-prod: ## 部署到生产环境
	@echo "🚀 部署到生产环境..."
	@echo "⚠️  生产环境部署需要额外权限确认"
	@# TODO: 实现生产环境部署逻辑

# ====================
# 开发工具
# ====================

install-tools: ## 安装开发工具
	@echo "🔧 安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ 开发工具安装完成"

db-shell: ## 连接数据库
	@echo "🗄️  连接数据库..."
	@$(DOCKER_COMPOSE) exec mysql mysql -u proxy_user -pproxy_pass123 proxy_platform

redis-shell: ## 连接Redis
	@echo "📦 连接Redis..."
	@$(DOCKER_COMPOSE) exec redis redis-cli

# ====================
# 性能测试
# ====================

load-test: ## 负载测试
	@echo "🔥 执行负载测试..."
	@# TODO: 实现负载测试脚本

benchmark: ## 性能基准测试
	@echo "⚡ 执行性能基准测试..."
	@$(GO_CMD) test -bench=. -benchmem ./... 