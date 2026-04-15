package postgres

import (
	"context"
	"time"

	"live-interact-engine/services/room-service/ent"
	entroom "live-interact-engine/services/room-service/ent/room"
	"live-interact-engine/services/room-service/internal/domain"

	"github.com/google/uuid"
)

// RoomRepository 实现 domain.RoomRepository 接口
type RoomRepository struct {
	client *ent.Client
}

// NewRoomRepository 创建 RoomRepository 实例
func NewRoomRepository(client *ent.Client) domain.RoomRepository {
	return &RoomRepository{
		client: client,
	}
}

// SaveRoom 保存房间信息（插入或更新）
func (r *RoomRepository) SaveRoom(ctx context.Context, room *domain.Room) error {
	err := r.client.Room.
		Create().
		SetID(room.RoomID).
		SetOwnerID(room.OwnerID).
		SetTitle(room.Title).
		SetDescription(room.Description).
		SetCreatedAt(room.CreatedAt.Unix()).
		SetUpdatedAt(room.UpdatedAt.Unix()).
		SetIsActive(room.IsActive).
		OnConflictColumns(entroom.FieldID).
		UpdateNewValues().
		Exec(ctx)

	return err
}

// GetRoom 获取房间信息
func (r *RoomRepository) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	entRoom, err := r.client.Room.Query().
		Where(entroom.IDEQ(roomID)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Room{
		RoomID:      entRoom.ID,
		OwnerID:     entRoom.OwnerID,
		Title:       entRoom.Title,
		Description: entRoom.Description,
		CreatedAt:   time.Unix(entRoom.CreatedAt, 0),
		UpdatedAt:   time.Unix(entRoom.UpdatedAt, 0),
		IsActive:    entRoom.IsActive,
	}, nil
}

// DeleteRoom 删除房间
func (r *RoomRepository) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	err := r.client.Room.
		DeleteOneID(roomID).
		Exec(ctx)

	if ent.IsNotFound(err) {
		return nil
	}

	return err
}
