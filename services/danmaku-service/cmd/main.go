package main

import (
	"context"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/grpc"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/subscription"
	"live-interact-engine/services/danmaku-service/internal/service"
	"live-interact-engine/shared/env"
	_ "live-interact-engine/shared/logger"
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
	// 初始化追踪
	tp, err := telemetry.InitTracer("danmaku-service")
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

	// 初始化依赖
	log.Println("Initializing dependencies...")
	subCfg := &subscription.ManagerConfig{
		Type: "memory",
	}

	danmakuService, err := service.NewDanmakuService(subCfg)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// 创建 gRPC 服务器
	handler := grpc.NewDanmakuHandler(danmakuService)
	opts := telemetry.SetupGRPCServerTracing()
	grpcServer := grpc_server.NewServer(opts...)

	// 注册服务
	pb.RegisterDanmakuServiceServer(grpcServer, handler)

	// 启动监听
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcAddr, err)
	}

	log.Printf("Danmaku service started, listening on %s", listener.Addr())

	// 启动服务器
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down danmaku service...")

	grpcServer.GracefulStop()
	log.Println("Danmaku service stopped")
}
