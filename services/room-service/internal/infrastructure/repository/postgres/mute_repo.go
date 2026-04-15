package postgres

import (
	"context"
	"time"

	"live-interact-engine/services/room-service/ent"
	entmute "live-interact-engine/services/room-service/ent/mute"
	"live-interact-engine/services/room-service/internal/domain"

	"github.com/google/uuid"
)

type MuteRepository struct {
	client *ent.Client
}

func NewMuteRepository(client *ent.Client) domain.MuteRepository {
	return &MuteRepository{client: client}
}

func (r *MuteRepository) Save(ctx context.Context, mute *domain.Mute) error {
	expiresAt := time.Unix(mute.MutedAt.Unix()+mute.Duration, 0)

	// 使用 room_id + user_id 作为冲突目标
	err := r.client.Mute.Create().
		SetRoomID(mute.RoomID).
		SetUserID(mute.UserID).
		SetAdminID(mute.AdminID).
		SetReason(mute.Reason).
		SetDuration(mute.Duration).
		SetMutedAt(mute.MutedAt.Unix()).
		SetExpiresAt(expiresAt.Unix()).
		OnConflictColumns(entmute.FieldRoomID, entmute.FieldUserID).
		UpdateNewValues().
		Exec(ctx)

	return err
}

func (r *MuteRepository) Delete(ctx context.Context, roomID, userID uuid.UUID) error {
	_, err := r.client.Mute.Delete().
		Where(
			entmute.RoomIDEQ(roomID),
			entmute.UserIDEQ(userID),
		).
		Exec(ctx)
	return err
}

func (r *MuteRepository) GetByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.Mute, error) {
	entMute, err := r.client.Mute.Query().
		Where(
			entmute.RoomIDEQ(roomID),
			entmute.UserIDEQ(userID),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Mute{
		ID:        entMute.ID,
		RoomID:    entMute.RoomID,
		UserID:    entMute.UserID,
		AdminID:   entMute.AdminID,
		Reason:    entMute.Reason,
		Duration:  entMute.Duration,
		MutedAt:   time.Unix(entMute.MutedAt, 0),
		ExpiresAt: time.Unix(entMute.ExpiresAt, 0),
		CreatedAt: time.Unix(entMute.CreatedAt, 0),
		UpdatedAt: time.Unix(entMute.UpdatedAt, 0),
	}, nil
}

func (r *MuteRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*domain.Mute, error) {
	mutes, err := r.client.Mute.Query().
		Where(entmute.RoomIDEQ(roomID)).
		Order(ent.Desc(entmute.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*domain.Mute, len(mutes))
	for i, m := range mutes {
		result[i] = &domain.Mute{
			ID:        m.ID,
			RoomID:    m.RoomID,
			UserID:    m.UserID,
			AdminID:   m.AdminID,
			Reason:    m.Reason,
			Duration:  m.Duration,
			MutedAt:   time.Unix(m.MutedAt, 0),
			ExpiresAt: time.Unix(m.ExpiresAt, 0),
			CreatedAt: time.Unix(m.CreatedAt, 0),
			UpdatedAt: time.Unix(m.UpdatedAt, 0),
		}
	}
	return result, nil
}

func (r *MuteRepository) DeleteExpired(ctx context.Context) (int, error) {
	now := time.Now().Unix()
	return r.client.Mute.Delete().
		Where(entmute.ExpiresAtLT(now)).
		Exec(ctx)
}
