package app

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	grpc_handler "live-interact-engine/services/user-service/internal/infrastructure/grpc"
	"live-interact-engine/services/user-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/user-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/user-service/internal/service"
	"log"

	"google.golang.org/grpc"
)

// Services 包含所有业务服务
type Services struct {
	UserService  domain.UserService
	TokenService domain.TokenService
}

// Handlers 包含所有 gRPC handlers
type Handlers struct {
	UserHandler       *grpc_handler.UserHandler
	TokenHandler      *grpc_handler.TokenHandler
	GiftServiceClient *grpc.ClientConn // gift-service gRPC 连接，用于优雅关闭
}

// InitDependencies 初始化所有依赖
func InitDependencies() (*Handlers, error) {
	ctx := context.Background()

	// ==================== 初始化 gRPC 客户端 ====================

	// 初始化 gift-service gRPC client
	giftWalletClient, giftClientConn, err := InitGiftServiceClient()
	if err != nil {
		log.Fatalf("初始化 gift-service 客户端失败: %v", err)
	}

	// ==================== 初始化 Repositories ====================

	// Ent Client
	client, err := postgres.NewEntClient(ctx)
	if err != nil {
		log.Fatalf("初始化 Ent Client 失败: %v", err)
	}

	// 用户 Repository
	userRepo := postgres.NewUserRepository(client)

	// Redis Token Repository
	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("初始化 Redis Client 失败: %v", err)
	}

	tokenRepo, err := redis.NewTokenRepository(redisClient)
	if err != nil {
		log.Fatalf("初始化 Redis TokenRepository 失败: %v", err)
	}

	// ==================== 初始化 Services ====================

	// 先初始化 UserService（带 giftWalletClient）
	userService, err := service.NewUserService(userRepo, giftWalletClient)
	if err != nil {
		log.Fatalf("初始化 UserService 失败: %v", err)
	}

	// 初始化 TokenService
	tokenService, err := service.NewTokenService(tokenRepo)
	if err != nil {
		log.Fatalf("初始化 TokenService 失败: %v", err)
	}

	// 通过 setter 注入 tokenService 到 userService（解决循环依赖）
	if userSvc, ok := userService.(*service.UserService); ok {
		userSvc.SetTokenService(tokenService)
	}

	// ==================== 初始化 Handlers ====================

	userHandler := grpc_handler.NewUserHandler(userService)
	tokenHandler := grpc_handler.NewTokenHandler(tokenService)

	return &Handlers{
		UserHandler:       userHandler,
		TokenHandler:      tokenHandler,
		GiftServiceClient: giftClientConn,
	}, nil
}
