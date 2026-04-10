package main

import (
	"context"
	"live-interact-engine/services/room-service/internal/app"
	"live-interact-engine/shared/env"
	pb "live-interact-engine/shared/proto/room"
	"live-interact-engine/shared/telemetry"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// 初始化追踪
	tp, err := telemetry.InitTracer("room-service")
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
	deps, err := app.InitDependencies(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// 启动 gRPC 服务器
	grpcPort := env.GetString("GRPC_PORT", ":9095")
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// 创建 gRPC 服务器
	opts := telemetry.SetupGRPCServerTracing()
	grpcServer := grpc.NewServer(opts...)

	// 注册服务
	pb.RegisterRoomServiceServer(grpcServer, deps.RoomHandler)

	log.Printf("Room service started, listening on %s", grpcPort)

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
	log.Println("Shutting down room service...")

	grpcServer.GracefulStop()
	log.Println("Room service stopped")
}
