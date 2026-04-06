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

	envFile := ".env." + mode

	// godotenv.Load(envFile) // 忽略错误，因为环境变量可能由容器设置
	if err := godotenv.Load(envFile); err != nil {
		if mode == "dev" {
			log.Panicf("加载 %s 失败, mode:%v, err: %v", envFile, err)
		}
		log.Printf("加载 %s 失败, mode:%v, err: %v", envFile, err)
	}
}

func main() {
	// 初始化追踪
	tp, err := telemetry.InitTracer("api-service")
	if err != nil {
		log.Fatalf("Tracer 初始化失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Tracer shutdown 失败: %v", err)
		}
	}()

	// 初始化 Metrics（API 服务才需要暴露 metrics）
	if err := telemetry.InitMetrics("api-service"); err != nil {
		log.Fatalf("Metrics 初始化失败: %v", err)
	}

	// 启动 Prometheus metrics 服务（独立端口）
	metricPort := env.GetString("METRICS_PORT", "9091")
	if err := telemetry.StartMetricsServer(metricPort); err != nil {
		log.Fatalf("Metrics 服务启动失败: %v", err)
	}

	// 从环境变量获取端口
	port := env.GetString("SERVER_PORT", "8080")

	// 启动 HTTP 服务器，指定路由注册函数
	server.RunHttpServer(
		port,
		router.RegisterRoutes,
		router.CloseClients,
	)
}
