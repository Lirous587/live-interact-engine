PROTO_DIR := proto
PROTO_SRC := \
		$(wildcard proto/danmaku/*.proto) \
		$(wildcard proto/user/*.proto) \
		$(wildcard proto/room/*.proto)
GO_OUT := .

.PHONY: generate-proto
generate-proto:
		protoc --proto_path=$(PROTO_DIR) --go_out=$(GO_OUT) --go-grpc_out=$(GO_OUT) $(PROTO_SRC)

.PHONY: build-services
build-services:
		powershell -ExecutionPolicy Bypass -File build.ps1

# 完整部署：编译 + 构建镜像 + 启动容器
.PHONY: up
up: build-services
		docker-compose up -d
		@echo "✓ Services started successfully!"
		@echo "- API Server: http://localhost:8080"
		@echo "- Prometheus: http://localhost:9091"
		@echo "- Grafana: http://localhost:3000"
		@echo "- Jaeger: http://localhost:16686"

# 停止容器
.PHONY: down
down:
		docker-compose down

# 查看日志
.PHONY: logs
logs:
		docker-compose logs -f api-service


# 完整重新部署
.PHONY: rebuild
rebuild: down build-services
		docker-compose up -d --build
		@echo "✓ Services rebuilt and started!"


# 快速重启
.PHONY: restart
restart:
		docker-compose restart
		@echo "✓ Services restarted!"

# 完全清理（删除所有，包括数据）
.PHONY: clean
clean:
		docker-compose down -v
		@echo "✓ All containers, networks, and volumes removed!"
