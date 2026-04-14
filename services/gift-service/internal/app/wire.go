package app

import (
	"context"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/events"
	"live-interact-engine/services/gift-service/internal/infrastructure/grpc"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/gift-service/internal/service"
	"live-interact-engine/shared/env"
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
	Publisher          *events.Publisher
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

	// ==================== 初始化 RabbitMQ Publisher ====================

	rabbitmqURL := env.GetString("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	publisher, err := events.NewPublisher(rabbitmqURL)
	if err != nil {
		log.Printf("警告: RabbitMQ Publisher 初始化失败: %v（将以非关键模式继续）", err)
		// 不中断启动，Publisher 为 nil 时处理
	}

	// ==================== 初始化 gRPC Handlers ====================
	giftHandler := grpc.NewGiftHandler(giftService, giftRecordService, publisher)
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
		Publisher:          publisher,
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
