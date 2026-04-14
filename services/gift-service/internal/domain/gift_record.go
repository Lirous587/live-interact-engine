package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GiftRecordStatus string

const (
	GiftRecordStatusPending GiftRecordStatus = "pending"
	GiftRecordStatusSuccess GiftRecordStatus = "success"
	GiftRecordStatusFailed  GiftRecordStatus = "failed"
)

// GiftRecord 送礼流水记录 聚合根
type GiftRecord struct {
	IdempotencyKey uuid.UUID
	UserID         uuid.UUID
	AnchorID       uuid.UUID
	RoomID         uuid.UUID
	GiftID         uuid.UUID
	Amount         int64
	Status         GiftRecordStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateGiftRecordRequest struct {
	IdempotencyKey uuid.UUID
	UserID         uuid.UUID
	AnchorID       uuid.UUID
	RoomID         uuid.UUID
	GiftID         uuid.UUID
	Amount         int64
}

// NewGiftRecord 从请求创建送礼记录
func NewGiftRecord(req *CreateGiftRecordRequest) (*GiftRecord, error) {
	if req.IdempotencyKey == uuid.Nil {
		return nil, ErrEmptyIdempotencyKey
	}
	if req.UserID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if req.AnchorID == uuid.Nil {
		return nil, ErrInvalidAnchorID
	}
	if req.RoomID == uuid.Nil {
		return nil, ErrInvalidRoomID
	}
	if req.GiftID == uuid.Nil {
		return nil, ErrInvalidGiftID
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &GiftRecord{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         req.UserID,
		AnchorID:       req.AnchorID,
		RoomID:         req.RoomID,
		GiftID:         req.GiftID,
		Amount:         req.Amount,
		Status:         GiftRecordStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// MarkAsSuccess 将记录标记为成功
func (gr *GiftRecord) MarkAsSuccess() error {
	if gr.Status != GiftRecordStatusPending {
		return ErrStatusTransition
	}
	gr.Status = GiftRecordStatusSuccess
	gr.UpdatedAt = time.Now()
	return nil
}

// MarkAsFailed 将记录标记为失败
func (gr *GiftRecord) MarkAsFailed() error {
	if gr.Status != GiftRecordStatusPending {
		return ErrStatusTransition
	}
	gr.Status = GiftRecordStatusFailed
	gr.UpdatedAt = time.Now()
	return nil
}

// IsPending 检查是否处于待处理状态
func (gr *GiftRecord) IsPending() bool {
	return gr.Status == GiftRecordStatusPending
}

// IsSuccess 检查是否成功
func (gr *GiftRecord) IsSuccess() bool {
	return gr.Status == GiftRecordStatusSuccess
}

type GiftRecordService interface {
	SaveGiftRecord(ctx context.Context, record *GiftRecord) error

	GetGiftRecordByKey(ctx context.Context, key uuid.UUID) (*GiftRecord, error)

	ListGiftRecordsByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*GiftRecord, error)

	ListGiftRecordsByAnchor(ctx context.Context, anchorID uuid.UUID, offset, limit int) ([]*GiftRecord, error)
}

// GiftRecordRepository 送礼流水仓储接口
type GiftRecordRepository interface {
	SaveGiftRecord(ctx context.Context, record *GiftRecord) error

	GetGiftRecordByKey(ctx context.Context, key uuid.UUID) (*GiftRecord, error)

	ListGiftRecordsByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*GiftRecord, error)

	ListGiftRecordsByAnchor(ctx context.Context, anchorID uuid.UUID, offset, limit int) ([]*GiftRecord, error)

	DeleteGiftRecordByKey(ctx context.Context, key uuid.UUID) error
}
