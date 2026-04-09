package main

import (
	"context"
	"live-interact-engine/services/user-service/internal/app"
	"live-interact-engine/shared/env"
	pb "live-interact-engine/shared/proto/user"
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
	grpcAddr = env.GetString("GRPC_ADDR", ":9094")
)

func main() {
	// 初始化追踪
	tp, err := telemetry.InitTracer("user-service")
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

	// 初始化所有依赖（repositories、services、handlers）
	handlers, err := app.InitDependencies()
	if err != nil {
		log.Fatalf("初始化依赖失败: %v", err)
	}

	// 创建 gRPC Server
	opts := telemetry.SetupGRPCServerTracing()
	grpcServer := grpc_server.NewServer(opts...)

	// 注册所有 gRPC 服务
	pb.RegisterUserServiceServer(grpcServer, handlers.UserHandler)
	pb.RegisterRoomAuthorizationServiceServer(grpcServer, handlers.RoomAuthorizationHandler)
	pb.RegisterTokenServiceServer(grpcServer, handlers.TokenHandler)

	// 启动监听
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	log.Printf("User gRPC 服务启动，监听端口 %s", lis.Addr())

	// 启动服务（后台运行）
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC 服务启动失败: %v", err)
		}
	}()

	// 监听 OS 信号（Ctrl+C、SIGTERM）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// 阻塞直到收到信号
	sig := <-sigChan
	log.Printf("收到信号: %v，开始优雅关闭...", sig)

	// 优雅关闭
	grpcServer.GracefulStop()
	log.Println("gRPC 服务已关闭")
}
