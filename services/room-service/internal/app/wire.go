package app

import (
	"context"
	"live-interact-engine/services/room-service/internal/domain"
	"live-interact-engine/services/room-service/internal/infrastructure/grpc"
	"live-interact-engine/services/room-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/room-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps 包含所有依赖
type Deps struct {
	RoomRepo         domain.RoomRepository
	UserRoomRoleRepo domain.UserRoomRoleRepository
	RoomService      domain.RoomService
	RoomHandler      *grpc.RoomHandler
}

// InitDependencies 初始化所有依赖
func InitDependencies(ctx context.Context) (*Deps, error) {
	// 初始化 PostgreSQL 连接池
	pool, err := postgres.NewPostgresDB(ctx)
	if err != nil {
		return nil, err
	}

	// 初始化 Repository
	roomRepo := postgres.NewRoomRepository(pool)
	userRoomRoleRepo := postgres.NewUserRoomRoleRepository(pool)

	// 初始化 Service
	roomService := service.NewRoomService(roomRepo, userRoomRoleRepo)

	// 初始化 gRPC Handler
	roomHandler := grpc.NewRoomHandler(roomService)

	return &Deps{
		RoomRepo:         roomRepo,
		UserRoomRoleRepo: userRoomRoleRepo,
		RoomService:      roomService,
		RoomHandler:      roomHandler,
	}, nil
}

// Close 关闭资源
func Close(pool *pgxpool.Pool) error {
	if pool != nil {
		pool.Close()
	}
	return nil
}
