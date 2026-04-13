package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Room 房间基础信息
type Room struct {
	RoomID      uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsActive    bool
}

type RoomService interface {
	CreateRoom(ctx context.Context, title, description string, ownerID uuid.UUID) (*Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*Room, error)
	AssignRole(ctx context.Context, ownerID, roomID, userID uuid.UUID, roleName string) error
	GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*UserRoomRole, error)
	CheckPermission(ctx context.Context, userID, roomID uuid.UUID, permission Permission) (bool, error)

	MuteUser(ctx context.Context, roomID, userID, adminID uuid.UUID, duration int64, reason string) error
	UnmuteUser(ctx context.Context, roomID, userID uuid.UUID) error
	IsMuted(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
	GetMuteInfo(ctx context.Context, roomID, userID uuid.UUID) (*Mute, error)
	GetMuteList(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*Mute, error)
}

type RoomRepository interface {
	// 保存房间
	SaveRoom(ctx context.Context, room *Room) error

	// 获取房间
	GetRoom(ctx context.Context, roomID uuid.UUID) (*Room, error)

	// 删除房间
	DeleteRoom(ctx context.Context, roomID uuid.UUID) error
}
