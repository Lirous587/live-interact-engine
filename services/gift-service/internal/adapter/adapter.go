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

// ==================== GiftRecord Adapters ====================

// GiftRecordToDomainPB 将 domain.GiftRecord 转换为 protobuf GiftRecord
func GiftRecordToDomainPB(record *domain.GiftRecord) *pb.GiftRecord {
	if record == nil {
		return nil
	}

	return &pb.GiftRecord{
		IdempotencyKey: record.IdempotencyKey.String(),
		UserId:         record.UserID.String(),
		AnchorId:       record.AnchorID.String(),
		RoomId:         record.RoomID.String(),
		GiftId:         record.GiftID.String(),
		Amount:         record.Amount,
		Status:         GiftRecordStatusToPB(record.Status),
		CreatedAt:      record.CreatedAt.Unix(),
		UpdatedAt:      record.UpdatedAt.Unix(),
	}
}

// GiftRecordStatusToPB 将 domain.GiftRecordStatus 转换为 protobuf GiftRecordStatus
func GiftRecordStatusToPB(status domain.GiftRecordStatus) pb.GiftRecordStatus {
	switch status {
	case domain.GiftRecordStatusPending:
		return pb.GiftRecordStatus_GIFT_RECORD_STATUS_PENDING
	case domain.GiftRecordStatusSuccess:
		return pb.GiftRecordStatus_GIFT_RECORD_STATUS_SUCCESS
	case domain.GiftRecordStatusFailed:
		return pb.GiftRecordStatus_GIFT_RECORD_STATUS_FAILED
	default:
		return pb.GiftRecordStatus_GIFT_RECORD_STATUS_UNSPECIFIED
	}
}

// ==================== LeaderboardEntry Adapters ====================

// LeaderboardEntryToDomainPB 将 domain 排行榜条目转换为 protobuf LeaderboardEntry
// TODO: 定义 domain.LeaderboardEntry 结构并完善此适配器
func LeaderboardEntryToDomainPB(userID string, score int64, rank int32) *pb.LeaderboardEntry {
	return &pb.LeaderboardEntry{
		UserId: userID,
		Score:  score,
		Rank:   rank,
	}
}
