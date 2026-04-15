package main

import (
	"context"
	"fmt"
	"live-interact-engine/services/gift-service/internal/app"
	"live-interact-engine/services/gift-service/internal/domain"
	"os"

	"log"
	"time"

	"github.com/google/uuid"
)

// GiftSeed 礼物初始化数据
type GiftSeed struct {
	ID          uuid.UUID
	Name        string
	Description string
	IconURL     string
	CacheKey    string
	Price       int64
	VIPOnly     bool
	Status      domain.GiftStatus
}

var gifts = []GiftSeed{
	// 普通礼物
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440001")),
		Name:        "点赞",
		Description: "最简单的表达方式",
		IconURL:     "https://example.com/thumbs-up.png",
		CacheKey:    "gift:thumbs_up",
		Price:       1,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440002")),
		Name:        "玫瑰花",
		Description: "传递爱意的礼物",
		IconURL:     "https://example.com/rose.png",
		CacheKey:    "gift:rose",
		Price:       10,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440003")),
		Name:        "钻戒",
		Description: "贵重的礼物",
		IconURL:     "https://example.com/diamond_ring.png",
		CacheKey:    "gift:diamond_ring",
		Price:       100,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440004")),
		Name:        "劳斯莱斯",
		Description: "极致奢华",
		IconURL:     "https://example.com/rolls_royce.png",
		CacheKey:    "gift:rolls_royce",
		Price:       1000,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440005")),
		Name:        "城堡",
		Description: "送你一座城堡",
		IconURL:     "https://example.com/castle.png",
		CacheKey:    "gift:castle",
		Price:       5000,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	// VIP专属礼物
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440010")),
		Name:        "VIP勋章",
		Description: "VIP用户专属",
		IconURL:     "https://example.com/vip_badge.png",
		CacheKey:    "gift:vip_badge",
		Price:       50,
		VIPOnly:     true,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440011")),
		Name:        "皇冠",
		Description: "VIP身份象征",
		IconURL:     "https://example.com/crown.png",
		CacheKey:    "gift:crown",
		Price:       200,
		VIPOnly:     true,
		Status:      domain.GiftStatusOnline,
	},
	// 限时礼物
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440020")),
		Name:        "新春红包",
		Description: "新年限定礼物",
		IconURL:     "https://example.com/red_envelope.png",
		CacheKey:    "gift:red_envelope",
		Price:       88,
		VIPOnly:     false,
		Status:      domain.GiftStatusLimitedTime,
	},
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440021")),
		Name:        "圣诞礼物",
		Description: "圣诞节限定",
		IconURL:     "https://example.com/christmas_gift.png",
		CacheKey:    "gift:christmas_gift",
		Price:       66,
		VIPOnly:     false,
		Status:      domain.GiftStatusLimitedTime,
	},
	// 下线礼物
	{
		ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440030")),
		Name:        "过期礼物",
		Description: "已下线的礼物",
		IconURL:     "https://example.com/expired.png",
		CacheKey:    "gift:expired",
		Price:       1,
		VIPOnly:     false,
		Status:      domain.GiftStatusOffline,
	},
}

func main() {
	log.Println("Starting gift seed initialization...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dns := "postgres://postgres:password@localhost:5433/gift_service?sslmode=disable"

	os.Setenv("DATABASE_DSN", dns)

	// 初始化依赖
	deps, err := app.InitDependencies(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// 初始化礼物
	count := 0
	for _, seed := range gifts {
		gift := &domain.Gift{
			ID:          seed.ID,
			Name:        seed.Name,
			Description: seed.Description,
			IconURL:     seed.IconURL,
			CacheKey:    seed.CacheKey,
			Price:       seed.Price,
			VIPOnly:     seed.VIPOnly,
			Status:      seed.Status,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := deps.GiftRepo.SaveGift(ctx, gift); err != nil {
			log.Printf("Failed to save gift %s: %v", seed.Name, err)
			continue
		}

		count++
		fmt.Printf("✓ Initialized gift: %s (ID: %s, Price: %d)\n", seed.Name, seed.ID.String(), seed.Price)
	}

	log.Printf("\nSuccessfully initialized %d gifts", count)
}
