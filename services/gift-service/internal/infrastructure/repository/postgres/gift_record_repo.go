package postgres

import (
	"context"

	"live-interact-engine/services/gift-service/ent"
	"live-interact-engine/services/gift-service/internal/domain"

	entrecord "live-interact-engine/services/gift-service/ent/giftrecord"

	"github.com/google/uuid"
)

type GiftRecordRepository struct {
	client *ent.Client
}

func NewGiftRecordRepository(client *ent.Client) domain.GiftRecordRepository {
	return &GiftRecordRepository{
		client: client,
	}
}

func (r *GiftRecordRepository) SaveGiftRecord(ctx context.Context, record *domain.GiftRecord) error {
	err := r.client.GiftRecord.Create().
		SetIdempotencyKey(record.IdempotencyKey).
		SetUserID(record.UserID).
		SetAnchorID(record.AnchorID).
		SetRoomID(record.RoomID).
		SetGiftID(record.GiftID).
		SetAmount(record.Amount).
		SetStatus(entrecord.Status(record.Status)).
		OnConflictColumns(entrecord.FieldIdempotencyKey).
		UpdateNewValues().
		Exec(ctx)

	return err
}

func (r *GiftRecordRepository) GetGiftRecordByKey(ctx context.Context, key uuid.UUID) (*domain.GiftRecord, error) {
	entRecord, err := r.client.GiftRecord.Query().
		Where(entrecord.IdempotencyKeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entGiftRecordToDomain(entRecord), nil
}

func (r *GiftRecordRepository) ListGiftRecordsByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*domain.GiftRecord, error) {
	entRecords, err := r.client.GiftRecord.Query().
		Where(entrecord.RoomIDEQ(roomID)).
		Order(entrecord.ByCreatedAt()).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	records := make([]*domain.GiftRecord, len(entRecords))
	for i, r := range entRecords {
		records[i] = entGiftRecordToDomain(r)
	}
	return records, nil
}

func (r *GiftRecordRepository) ListGiftRecordsByAnchor(ctx context.Context, anchorID uuid.UUID, offset, limit int) ([]*domain.GiftRecord, error) {
	entRecords, err := r.client.GiftRecord.Query().
		Where(entrecord.AnchorIDEQ(anchorID)).
		Order(entrecord.ByCreatedAt()).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	records := make([]*domain.GiftRecord, len(entRecords))
	for i, r := range entRecords {
		records[i] = entGiftRecordToDomain(r)
	}
	return records, nil
}

func (r *GiftRecordRepository) DeleteGiftRecordByKey(ctx context.Context, key uuid.UUID) error {
	count, err := r.client.GiftRecord.Delete().Where(entrecord.IdempotencyKeyEQ(key)).Exec(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func entGiftRecordToDomain(entRecord *ent.GiftRecord) *domain.GiftRecord {
	return &domain.GiftRecord{
		IdempotencyKey: entRecord.IdempotencyKey,
		UserID:         entRecord.UserID,
		AnchorID:       entRecord.AnchorID,
		RoomID:         entRecord.RoomID,
		GiftID:         entRecord.GiftID,
		Amount:         entRecord.Amount,
		Status:         domain.GiftRecordStatus(entRecord.Status),
		CreatedAt:      entRecord.CreatedAt,
		UpdatedAt:      entRecord.UpdatedAt,
	}
}
