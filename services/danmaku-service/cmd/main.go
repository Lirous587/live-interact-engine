package main

import (
	"context"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/grpc"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/subscription"
	"live-interact-engine/services/danmaku-service/internal/service"
	"live-interact-engine/shared/env"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/telemetry"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_server "google.golang.org/grpc"
)

var (
	grpcAddr = env.GetString("GRPC_ADDR", ":9093")
)

func main() {
	// 初始化 OTel
	otelProviders, err := telemetry.InitOTelProviders("danmaku-service")
	if err != nil {
		log.Fatalf("OTel 初始化失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := otelProviders.Shutdown(ctx); err != nil {
			log.Printf("OTel shutdown 失败: %v", err)
		}
	}()

	// 1. 创建 SubscriptionManager
	subCfg := &subscription.ManagerConfig{
		Type: "memory",
	}

	// 2. 创建 Service
	danmakuService, err := service.NewDanmakuService(subCfg)
	if err != nil {
		log.Fatalf("创建 DanmakuService 失败: %v", err)
	}

	// 3. 创建 Handler
	handler := grpc.NewDanmakuHandler(danmakuService)

	// 4. 创建 gRPC Server
	opts := telemetry.SetupGRPCServerTracing()
	grpcServer := grpc_server.NewServer(opts...)

	// 5. 注册服务
	pb.RegisterDanmakuServiceServer(grpcServer, handler)

	// 6. 启动监听
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	log.Printf("Danmaku gRPC 服务启动，监听端口 :%s", lis.Addr())

	// 7. 启动服务（阻塞）
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC 服务启动失败: %v", err)
		}
	}()

	// 8. 监听 OS 信号（Ctrl+C、SIGTERM）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// 阻塞直到收到信号
	sig := <-sigChan
	log.Printf("收到信号: %v，开始优雅关闭...", sig)

	// 9. 优雅关闭
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	log.Println("gRPC 服务已关闭")
}
