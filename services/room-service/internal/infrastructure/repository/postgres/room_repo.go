package postgres

import (
	"context"
	"live-interact-engine/services/room-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RoomRepository 实现 domain.RoomRepository 接口
type RoomRepository struct {
	pool *pgxpool.Pool
}

// NewRoomRepository 创建 RoomRepository 实例
func NewRoomRepository(pool *pgxpool.Pool) domain.RoomRepository {
	return &RoomRepository{
		pool: pool,
	}
}

// SaveRoom 保存房间信息
func (r *RoomRepository) SaveRoom(ctx context.Context, room *domain.Room) error {
	sql := `
		INSERT INTO rooms (room_id, owner_id, title, description, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (room_id) DO UPDATE SET
			title = $3,
			description = $4,
			updated_at = $6,
			is_active = $7
	`

	_, err := r.pool.Exec(ctx, sql,
		room.RoomID,
		room.OwnerID,
		room.Title,
		room.Description,
		room.CreatedAt.Unix(),
		room.UpdatedAt.Unix(),
		room.IsActive,
	)

	return err
}

// GetRoom 获取房间信息
func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	sql := `
		SELECT room_id, owner_id, title, description, created_at, updated_at, is_active
		FROM rooms
		WHERE room_id = $1
	`

	row := r.pool.QueryRow(ctx, sql, roomID)

	var room domain.Room
	err := row.Scan(
		&room.RoomID,
		&room.OwnerID,
		&room.Title,
		&room.Description,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.IsActive,
	)

	if err != nil {
		return nil, err
	}

	// Convert Unix timestamp back to time.Time
	room.CreatedAt = room.CreatedAt
	room.UpdatedAt = room.UpdatedAt

	return &room, nil
}

// DeleteRoom 删除房间
func (r *RoomRepository) DeleteRoom(ctx context.Context, roomID string) error {
	sql := `DELETE FROM rooms WHERE room_id = $1`
	_, err := r.pool.Exec(ctx, sql, roomID)
	return err
}
