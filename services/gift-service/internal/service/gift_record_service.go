package service

import (
	"context"

	"github.com/google/uuid"

	"live-interact-engine/services/gift-service/internal/domain"
)

type GiftRecordService struct {
	recordRepo domain.GiftRecordRepository
}

func NewGiftRecordService(
	recordRepo domain.GiftRecordRepository,
	giftRepo domain.GiftRepository,
	walletSvc *WalletService,
) domain.GiftRecordService {
	return &GiftRecordService{
		recordRepo: recordRepo,
	}
}

func (s *GiftRecordService) SaveGiftRecord(ctx context.Context, record *domain.GiftRecord) error {
	return s.recordRepo.SaveGiftRecord(ctx, record)
}

func (s *GiftRecordService) GetGiftRecordByKey(ctx context.Context, key uuid.UUID) (*domain.GiftRecord, error) {
	return s.recordRepo.GetGiftRecordByKey(ctx, key)
}

func (s *GiftRecordService) ListGiftRecordsByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*domain.GiftRecord, error) {
	return s.recordRepo.ListGiftRecordsByRoom(ctx, roomID, offset, limit)
}

func (s *GiftRecordService) ListGiftRecordsByAnchor(ctx context.Context, anchorID uuid.UUID, offset, limit int) ([]*domain.GiftRecord, error) {
	return s.recordRepo.ListGiftRecordsByAnchor(ctx, anchorID, offset, limit)
}
