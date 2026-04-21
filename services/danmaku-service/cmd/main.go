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
	"live-interact-engine/shared/env"
	_ "live-interact-engine/shared/logger"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/telemetry"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

var (
	grpcAddr = env.GetString("GRPC_ADDR", ":9093")
	natsURL  = env.GetString("NATS_URL", nats.DefaultURL)
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

	// ==================== 初始化 Redis（历史回放 + 限流）====================
	log.Println("Connecting to Redis...")
	rdb, err := redisrepo.NewClient()
	if err != nil {
		log.Fatalf("Redis unavailable: %v", err)
	}

	// ==================== 初始化 NATS（跨节点广播，必选依赖）====================
	log.Printf("Connecting to NATS at %s...", natsURL)
	nc, err := nats.Connect(natsURL,
		nats.Name("danmaku-service"),
		nats.MaxReconnects(-1),           // 无限重连
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		log.Fatalf("NATS unavailable: %v", err)
	}
	defer nc.Drain()

	// ==================== 初始化订阅管理器 ====================
	subManager, err := subscription.NewNATSManager(nc)
	if err != nil {
		log.Fatalf("subscription manager init failed: %v", err)
	}
	defer subManager.Close()

	// ==================== 初始化历史回放 & 限流器 ====================
	danmakuHistory := redisrepo.NewDanmakuHistory(rdb)
	rateLimiter := ratelimit.NewDanmakuRateLimiter(rdb)

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
