package adapter

import (
	"live-interact-engine/services/gift-service/internal/domain"
	pb "live-interact-engine/shared/proto/gift"
)

// ==================== Gift Adapters ====================

// GiftToDomainPB 将 domain.Gift 转换为 protobuf Gift
func GiftToDomainPB(gift *domain.Gift) *pb.Gift {
	if gift == nil {
		return nil
	}

	return &pb.Gift{
		Id:          gift.ID.String(), // 假设 ID 是 uuid.UUID，需要转换为 int64 或字符串
		Name:        gift.Name,
		Description: gift.Description,
		IconUrl:     gift.IconURL,
		CacheKey:    gift.CacheKey,
		Price:       gift.Price,
		VipOnly:     gift.VIPOnly,
		Status:      GiftStatusToPB(gift.Status),
		CreatedAt:   gift.CreatedAt.Unix(),
		UpdatedAt:   gift.UpdatedAt.Unix(),
	}
}

// GiftStatusToPB 将 domain.GiftStatus 转换为 protobuf GiftStatus
func GiftStatusToPB(status domain.GiftStatus) pb.GiftStatus {
	switch status {
	case domain.GiftStatusOnline:
		return pb.GiftStatus_GIFT_STATUS_ONLINE
	case domain.GiftStatusOffline:
		return pb.GiftStatus_GIFT_STATUS_OFFLINE
	case domain.GiftStatusLimitedTime:
		return pb.GiftStatus_GIFT_STATUS_LIMITED_TIME
	default:
		return pb.GiftStatus_GIFT_STATUS_UNSPECIFIED
	}
}

// ==================== Wallet Adapters ====================

// WalletToDomainPB 将 domain.Wallet 转换为 protobuf Wallet
func WalletToDomainPB(wallet *domain.Wallet) *pb.Wallet {
	if wallet == nil {
		return nil
	}

	return &pb.Wallet{
		UserId:        wallet.UserID.String(),
		Balance:       wallet.Balance,
		VersionNumber: wallet.VersionNumber,
		CreatedAt:     wallet.CreatedAt.Unix(),
		UpdatedAt:     wallet.UpdatedAt.Unix(),
	}
}
