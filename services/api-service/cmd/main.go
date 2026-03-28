package main

import (
	"context"
	"live-interact-engine/services/api-service/internal/router"
	"live-interact-engine/services/api-service/internal/utils/server"
	"live-interact-engine/shared/env"
	"live-interact-engine/shared/telemetry"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	_ "live-interact-engine/shared/logger"
)

func init() {
	// 根据环境加载对应的 .env 文件
	mode := os.Getenv("SERVER_MODE")
	if mode == "" {
		mode = "dev"
	}

	envFile := ".env"
	if mode != "dev" {
		envFile = ".env." + mode
	}

	godotenv.Load(envFile) // 忽略错误，因为环境变量可能由容器设置
}

func main() {
	// 初始化 OTel providers
	otelProviders, err := telemetry.InitOTelProviders("api-service")
	if err != nil {
		log.Fatalf("OTel providers 初始化失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := otelProviders.Shutdown(ctx); err != nil {
			log.Printf("OTel shutdown 失败: %v", err)
		}
	}()

	// 启动 Prometheus metrics 服务（独立端口）
	if err := telemetry.StartMetricsServer("9090"); err != nil {
		log.Fatalf("Metrics 服务启动失败: %v", err)
	}

	// 从环境变量获取端口
	port := env.GetString("SERVER_PORT", "8080")

	// 启动 HTTP 服务器，指定路由注册函数
	server.RunHttpServer(
		port,
		router.RegisterRoutes,
		nil,
	)
}
