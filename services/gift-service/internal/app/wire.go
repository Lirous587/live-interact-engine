package app

import (
	"context"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/events"
	"live-interact-engine/services/gift-service/internal/infrastructure/filter/memory"
	"live-interact-engine/services/gift-service/internal/infrastructure/grpc"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/postgres"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/redis"
	"live-interact-engine/services/gift-service/internal/service"
	"live-interact-engine/shared/env"
	"log"
	"time"
)

// Deps 包含所有依赖
type Deps struct {
	GiftRepo              domain.GiftRepository
	WalletRepo            domain.WalletRepository
	WalletTransactionRepo domain.WalletTransactionRepository
	GiftCache             domain.GiftCache
	WalletCache           domain.WalletCache
	GiftService           *service.GiftService
	WalletService         *service.WalletService
	Publisher             *events.Publisher
	Consumer              *events.Consumer
	GiftHandler           *grpc.GiftHandler
	WalletHandler         *grpc.WalletHandler
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
	walletTransactionRepo := postgres.NewWalletTransactionRepository(entClient)

	// ==================== 初始化 Caches ====================
	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("初始化 Redis Client 失败: %v", err)
	}

	giftCache := redis.NewGiftCache(redisClient)
	walletCache := redis.NewWalletCache(redisClient)
	walletFilter := memory.NewLocalWalletFilter(time.Hour * 1)

	// ==================== 初始化 Services ====================
	giftService := service.NewGiftService(giftRepo, giftCache)
	walletService := service.NewWalletService(walletRepo, walletCache, walletFilter)

	// ==================== 初始化 RabbitMQ ====================
	rabbitmqURL := env.GetString("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// Publisher
	publisher, err := events.NewPublisher(rabbitmqURL)
	if err != nil {
		log.Printf("警告: RabbitMQ Publisher 初始化失败: %v（将以非关键模式继续）", err)
	}

	// Consumer
	consumer, err := events.NewConsumer(
		rabbitmqURL,
		walletTransactionRepo,
		walletRepo,
		walletService,
	)
	if err != nil {
		log.Printf("警告: RabbitMQ Consumer 初始化失败: %v", err)
		consumer = nil
	}

	// ==================== 初始化 gRPC Handlers ====================
	giftHandler := grpc.NewGiftHandler(giftService, walletService, publisher)
	walletHandler := grpc.NewWalletHandler(walletService, publisher)

	return &Deps{
		GiftRepo:              giftRepo,
		WalletRepo:            walletRepo,
		WalletTransactionRepo: walletTransactionRepo,
		GiftCache:             giftCache,
		WalletCache:           walletCache,
		GiftService:           giftService,
		WalletService:         walletService,
		Publisher:             publisher,
		Consumer:              consumer,
		GiftHandler:           giftHandler,
		WalletHandler:         walletHandler,
	}, nil
}

// Close 关闭资源
func Close(deps *Deps) error {
	if deps.Consumer != nil {
		deps.Consumer.Close()
	}
	if deps.Publisher != nil {
		deps.Publisher.Close()
	}
	return nil
}
