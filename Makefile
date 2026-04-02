PROTO_DIR := proto
PROTO_SRC := \
    $(wildcard proto/danmaku/*.proto) \
    $(wildcard proto/payment/*.proto) \
    $(wildcard proto/room/*.proto)
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc --proto_path=$(PROTO_DIR) --go_out=$(GO_OUT) --go-grpc_out=$(GO_OUT) $(PROTO_SRC)

.PHONY: build-api-service
build-api-service:
	powershell -ExecutionPolicy Bypass -File build.ps1

# 构建 Docker 镜像
.PHONY: docker-build
docker-build: build-api-service
	docker build -t api-service:latest .

# 启动 docker-compose
.PHONY: docker-up
docker-up: docker-build
	docker-compose up -d

# 停止 docker-compose
.PHONY: docker-down
docker-down:
	docker-compose down

# 完整的构建和启动（build + docker-compose build + 启动）
.PHONY: docker-run
docker-run: build-api-service
	docker-compose build
	docker-compose up -d
	@echo "Containers started successfully!"
	@echo "- API Server: http://localhost:8080"
	@echo "- Prometheus: http://localhost:9091"
	@echo "- Grafana: http://localhost:3000"
	@echo "- Jaeger: http://localhost:16686"

# 查看日志
.PHONY: docker-logs
docker-logs:
	docker-compose logs -f api-service