PROTO_DIR := proto
PROTO_SRC := \
		$(wildcard proto/danmaku/*.proto) \
		$(wildcard proto/user/*.proto) \
		$(wildcard proto/room/*.proto) \
		$(wildcard proto/gift/*.proto) 
GO_OUT := .

.PHONY: generate-proto
generate-proto:
		protoc --proto_path=$(PROTO_DIR) --go_out=$(GO_OUT) --go-grpc_out=$(GO_OUT) $(PROTO_SRC)

.PHONY: build-services
build-services:
		powershell -ExecutionPolicy Bypass -File build.ps1

.PHONY: generate-api-docs
generate-api-docs:
		swag init -g .\services\api-service\cmd\main.go -o api/openapi

.PHONY: gen-ent
gen-ent:
		cd services/user-service && go generate ./ent
		cd services/room-service && go generate ./ent
		cd services/gift-service && go generate ./ent

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

# 单个服务重新构建脚本
.PHONY: rebuild-api
rebuild-api: build-services
		docker-compose up -d --build api-service
		@echo "✓ api-service rebuilt!"

.PHONY: rebuild-user
rebuild-user: build-services
		docker-compose up -d --build user-service
		@echo "✓ user-service rebuilt!"

.PHONY: rebuild-room
rebuild-room: build-services
		docker-compose up -d --build room-service
		@echo "✓ room-service rebuilt!"

.PHONY: rebuild-danmaku
rebuild-danmaku: build-services
		docker-compose up -d --build danmaku-service
		@echo "✓ danmaku-service rebuilt!"

.PHONY: rebuild-gift
rebuild-gift: build-services
		docker-compose up -d --build gift-service
		@echo "✓ gift-service rebuilt!"

# 初始化礼物 seed 数据（需服务已启动，PostgreSQL 可访问）
.PHONY: seed-gifts
seed-gifts:
		powershell -Command "$$env:DATABASE_DSN='postgres://postgres:password@localhost:5433/gift_service?sslmode=disable'; $$env:ENT_MODE='prod'; go run ./services/gift-service/cmd/seed"
		@echo "✓ 礼物数据初始化完成！"

# 完全清理（删除所有，包括数据）
.PHONY: clean
clean:
		docker-compose down -v
		@echo "✓ All containers, networks, and volumes removed!"
