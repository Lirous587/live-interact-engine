package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Mute struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	UserID    uuid.UUID
	AdminID   uuid.UUID
	Reason    string
	Duration  int64
	MutedAt   time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Mute) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

type MuteRepository interface {
	Create(ctx context.Context, mute *Mute) error
	Delete(ctx context.Context, roomID, userID uuid.UUID) error
	GetByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*Mute, error)
	ListByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*Mute, error)
	DeleteExpired(ctx context.Context) (int, error)
}
