package postgres

import (
	"context"
	"encoding/json"
	"live-interact-engine/services/user-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRoomRoleRepository struct {
	pool *pgxpool.Pool
}

func NewUserRoomRoleRepository(pool *pgxpool.Pool) domain.UserRoomRoleRepository {
	return &UserRoomRoleRepository{pool: pool}
}

// GetUserRoomRole 获取用户在某房间的角色信息
func (r *UserRoomRoleRepository) GetUserRoomRole(ctx context.Context, userID, roomID string) (*domain.UserRoomRole, error) {
	query := `
		SELECT user_id, room_id, role_name, is_owner, permissions, created_at, updated_at
		FROM user_room_roles
		WHERE user_id = $1 AND room_id = $2
	`

	var urr domain.UserRoomRole
	var permissionsJSON []byte
	var createdAt, updatedAt int64

	err := r.pool.QueryRow(ctx, query, userID, roomID).Scan(
		&urr.UserID,
		&urr.RoomID,
		&urr.RoleName,
		&urr.IsOwner,
		&permissionsJSON,
		&createdAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 反序列化权限
	var permissionInts []int32
	if err := json.Unmarshal(permissionsJSON, &permissionInts); err == nil {
		urr.Permissions = make([]domain.Permission, len(permissionInts))
		for i, p := range permissionInts {
			urr.Permissions[i] = domain.Permission(p)
		}
	}

	urr.CreatedAt = time.Unix(createdAt, 0)
	urr.UpdatedAt = time.Unix(updatedAt, 0)

	return &urr, nil
}

// SaveUserRoomRole 保存用户房间角色
func (r *UserRoomRoleRepository) SaveUserRoomRole(ctx context.Context, urr *domain.UserRoomRole) error {
	// 序列化权限
	permissionInts := make([]int32, len(urr.Permissions))
	for i, p := range urr.Permissions {
		permissionInts[i] = int32(p)
	}
	permissionsJSON, err := json.Marshal(permissionInts)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_room_roles (user_id, room_id, role_name, is_owner, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, room_id) DO UPDATE SET
			role_name = $3,
			is_owner = $4,
			permissions = $5,
			updated_at = $7
	`

	_, err = r.pool.Exec(ctx, query,
		urr.UserID,
		urr.RoomID,
		urr.RoleName,
		urr.IsOwner,
		permissionsJSON,
		urr.CreatedAt.Unix(),
		urr.UpdatedAt.Unix(),
	)

	return err
}

// DeleteUserRoomRole 删除用户房间角色
func (r *UserRoomRoleRepository) DeleteUserRoomRole(ctx context.Context, userID, roomID string) error {
	query := `DELETE FROM user_room_roles WHERE user_id = $1 AND room_id = $2`

	_, err := r.pool.Exec(ctx, query, userID, roomID)
	return err
}
