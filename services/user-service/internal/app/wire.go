package app

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	grpc_handler "live-interact-engine/services/user-service/internal/infrastructure/grpc"
	"live-interact-engine/services/user-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/user-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/user-service/internal/service"
	"log"
)

// Services 包含所有业务服务
type Services struct {
	UserService              domain.UserService
	RoomAuthorizationService domain.RoomAuthorizationService
	TokenService             domain.TokenService
}

// Handlers 包含所有 gRPC handlers
type Handlers struct {
	UserHandler              *grpc_handler.UserHandler
	RoomAuthorizationHandler *grpc_handler.RoomAuthorizationHandler
	TokenHandler             *grpc_handler.TokenHandler
}

// InitDependencies 初始化所有依赖
func InitDependencies() (*Handlers, error) {
	ctx := context.Background()

	// ==================== 初始化 Repositories ====================

	// PostgreSQL 连接池
	pool, err := postgres.NewPostgresDB(ctx)
	if err != nil {
		log.Fatalf("初始化 PostgreSQL 失败: %v", err)
	}

	// 用户 Repository
	userRepo := postgres.NewUserRepository(pool)

	// 房间角色 Repository
	roomRoleRepo := postgres.NewUserRoomRoleRepository(pool)

	// Redis Token Repository
	tokenRepo, err := redis.NewTokenRepository()
	if err != nil {
		log.Fatalf("初始化 Redis TokenRepository 失败: %v", err)
	}

	// ==================== 初始化 Services ====================

	// 先初始化 UserService（不需要 tokenService）
	userService, err := service.NewUserService(userRepo)
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

	// 初始化 RoomAuthorizationService
	roomAuthService, err := service.NewRoomAuthorizationService(roomRoleRepo)
	if err != nil {
		log.Fatalf("初始化 RoomAuthorizationService 失败: %v", err)
	}

	// ==================== 初始化 Handlers ====================

	userHandler := grpc_handler.NewUserHandler(userService)
	roomAuthHandler := grpc_handler.NewRoomAuthorizationHandler(roomAuthService)
	tokenHandler := grpc_handler.NewTokenHandler(tokenService)

	return &Handlers{
		UserHandler:              userHandler,
		RoomAuthorizationHandler: roomAuthHandler,
		TokenHandler:             tokenHandler,
	}, nil
}
