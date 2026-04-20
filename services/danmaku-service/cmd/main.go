package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	svcgrpc "live-interact-engine/services/danmaku-service/internal/infrastructure/grpc"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/ratelimit"
	redisrepo "live-interact-engine/services/danmaku-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/subscription"
	"live-interact-engine/services/danmaku-service/internal/service"
	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/shared/env"
	_ "live-interact-engine/shared/logger"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/telemetry"

	"google.golang.org/grpc"
)

var (
	grpcAddr = env.GetString("GRPC_ADDR", ":9093")
)

func main() {
	// ==================== 初始化追踪 ====================
	tp, err := telemetry.InitTracer("danmaku-service")
	if err != nil {
		log.Fatalf("tracer init failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("tracer shutdown error: %v", err)
		}
	}()

	// ==================== 初始化 Redis ====================
	// Redis 为可选依赖：连接失败时降级为纯内存模式（单实例 / 开发环境）
	log.Println("Initializing Redis client...")
	rdb, redisErr := redisrepo.NewClient()
	if redisErr != nil {
		log.Printf("Warning: Redis unavailable (%v), falling back to memory mode", redisErr)
	}

	// ==================== 初始化订阅管理器 ====================
	// 环境变量 SUBSCRIPTION_TYPE=redis 使用跨节点 Pub/Sub；默认 memory
	subManager, err := subscription.NewManager(rdb)
	if err != nil {
		log.Fatalf("failed to init subscription manager: %v", err)
	}
	defer subManager.Close()

	// ==================== 初始化历史回放 & 限流器 ====================
	// 两者均依赖 Redis；Redis 不可用时置 nil，service 层会安全跳过（fail-open）
	var danmakuHistory domain.DanmakuHistory
	var rateLimiter domain.RateLimiter

	if redisErr == nil {
		danmakuHistory = redisrepo.NewDanmakuHistory(rdb)
		rateLimiter = ratelimit.NewDanmakuRateLimiter(rdb)
		log.Println("History replay and rate limiting enabled")
	} else {
		log.Println("Warning: history replay and rate limiting disabled (no Redis)")
	}

	// ==================== 初始化 Service & Handler ====================
	danmakuSvc := service.NewDanmakuService(subManager, danmakuHistory, rateLimiter)
	handler := svcgrpc.NewDanmakuHandler(danmakuSvc)

	// ==================== 启动 gRPC 服务器 ====================
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
	}

	opts := telemetry.SetupGRPCServerTracing()
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDanmakuServiceServer(grpcServer, handler)

	log.Printf("Danmaku service started, listening on %s", listener.Addr())

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// ==================== 优雅退出 ====================
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down danmaku service...")
	grpcServer.GracefulStop()
	log.Println("Danmaku service stopped")
}
