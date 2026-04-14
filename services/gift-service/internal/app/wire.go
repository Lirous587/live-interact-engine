package app

import (
	"context"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/grpc"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/gift-service/internal/service"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps 包含所有依赖
type Deps struct {
	GiftRepo           domain.GiftRepository
	WalletRepo         domain.WalletRepository
	GiftRecordRepo     domain.GiftRecordRepository
	GiftCache          domain.GiftCache
	WalletCache        domain.WalletCache
	GiftService        *service.GiftService
	WalletService      *service.WalletService
	GiftRecordService  domain.GiftRecordService
	GiftHandler        *grpc.GiftHandler
	GiftRecordHandler  *grpc.GiftRecordHandler
	WalletHandler      *grpc.WalletHandler
	LeaderboardHandler *grpc.LeaderboardHandler
}

// InitDependencies 初始化所有依赖
func InitDependencies(ctx context.Context) (*Deps, error) {
	// ==================== 初始化 Repositories ====================
	entClient, err := postgres.NewEntClient(ctx)
	if err != nil {
		log.Fatalf("初始化 Ent Client 失败: %v", err)
	}

	giftRepo := postgres.NewGiftRepository(entClient)
	walletRepo := postgres.NewWalletRepository(entClient)
	giftRecordRepo := postgres.NewGiftRecordRepository(entClient)

	// ==================== 初始化 Caches ====================
	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("初始化 Redis Client 失败: %v", err)
	}

	giftCache := redis.NewGiftCache(redisClient)
	walletCache := redis.NewWalletCache(redisClient)

	// ==================== 初始化 Services ====================
	giftService := service.NewGiftService(giftRepo, giftCache)
	walletService := service.NewWalletService(walletRepo, walletCache)
	giftRecordService := service.NewGiftRecordService(giftRecordRepo, giftRepo, walletService)

	// ==================== 初始化 gRPC Handlers ====================
	giftHandler := grpc.NewGiftHandler(giftService, giftRecordService)
	giftRecordHandler := grpc.NewGiftRecordHandler(giftRecordService)
	walletHandler := grpc.NewWalletHandler(walletService)
	leaderboardHandler := grpc.NewLeaderboardHandler()

	return &Deps{
		GiftRepo:           giftRepo,
		WalletRepo:         walletRepo,
		GiftRecordRepo:     giftRecordRepo,
		GiftCache:          giftCache,
		WalletCache:        walletCache,
		GiftService:        giftService,
		WalletService:      walletService,
		GiftRecordService:  giftRecordService,
		GiftHandler:        giftHandler,
		GiftRecordHandler:  giftRecordHandler,
		WalletHandler:      walletHandler,
		LeaderboardHandler: leaderboardHandler,
	}, nil
}

// Close 关闭资源
func Close(pool *pgxpool.Pool) error {
	if pool != nil {
		pool.Close()
	}
	return nil
}
