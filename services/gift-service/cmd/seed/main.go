package main

import (
	"context"
	"log"

	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/repository/postgres"

	"github.com/google/uuid"
)

// 固定 UUID 保证 seed 幂等（多次执行 upsert 结果一致）
var gifts = []*domain.Gift{
	{
		ID:          uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:        "玫瑰",
		Description: "送出你的爱意",
		IconURL:     "/icons/rose.png",
		CacheKey:    "rose",
		Price:       1,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.MustParse("00000001-0000-0000-0000-000000000002"),
		Name:        "棒棒糖",
		Description: "甜甜的支持",
		IconURL:     "/icons/lollipop.png",
		CacheKey:    "lollipop",
		Price:       5,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.MustParse("00000001-0000-0000-0000-000000000003"),
		Name:        "烟花",
		Description: "绚烂的祝福",
		IconURL:     "/icons/fireworks.png",
		CacheKey:    "fireworks",
		Price:       10,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.MustParse("00000001-0000-0000-0000-000000000004"),
		Name:        "超级星星",
		Description: "你是最亮的星",
		IconURL:     "/icons/star.png",
		CacheKey:    "super_star",
		Price:       50,
		VIPOnly:     false,
		Status:      domain.GiftStatusOnline,
	},
	{
		ID:          uuid.MustParse("00000001-0000-0000-0000-000000000005"),
		Name:        "豪华游艇",
		Description: "尊贵专属礼遇",
		IconURL:     "/icons/yacht.png",
		CacheKey:    "yacht",
		Price:       500,
		VIPOnly:     true,
		Status:      domain.GiftStatusOnline,
	},
}

func main() {
	ctx := context.Background()

	client, err := postgres.NewEntClient(ctx)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	giftRepo := postgres.NewGiftRepository(client)

	for _, g := range gifts {
		if err := giftRepo.SaveGift(ctx, g); err != nil {
			log.Fatalf("保存礼物 [%s] 失败: %v", g.Name, err)
		}
		log.Printf("  ✓ %-12s price=%-5d vip_only=%v", g.Name, g.Price, g.VIPOnly)
	}

	log.Printf("礼物数据初始化完成，共 %d 条", len(gifts))
}
