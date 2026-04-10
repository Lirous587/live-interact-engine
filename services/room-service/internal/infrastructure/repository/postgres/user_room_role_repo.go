package postgres

import (
	"context"
	"encoding/json"
	"live-interact-engine/services/room-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRoomRoleRepository 实现 domain.UserRoomRoleRepository 接口
type UserRoomRoleRepository struct {
	pool *pgxpool.Pool
}

// NewUserRoomRoleRepository 创建 UserRoomRoleRepository 实例
func NewUserRoomRoleRepository(pool *pgxpool.Pool) domain.UserRoomRoleRepository {
	return &UserRoomRoleRepository{
		pool: pool,
	}
}

// GetUserRoomRole 获取用户在房间的角色信息
func (r *UserRoomRoleRepository) GetUserRoomRole(ctx context.Context, userID, roomID string) (*domain.UserRoomRole, error) {
	sql := `
		SELECT user_id, room_id, role_name, permissions, created_at, updated_at
		FROM user_room_roles
		WHERE user_id = $1 AND room_id = $2
	`

	row := r.pool.QueryRow(ctx, sql, userID, roomID)

	var urr domain.UserRoomRole
	var permissionsJSON []byte

	err := row.Scan(
		&urr.UserID,
		&urr.RoomID,
		&urr.RoleName,
		&permissionsJSON,
		&urr.CreatedAt,
		&urr.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// 反序列化 JSONB 权限
	var permissions []int32
	if err := json.Unmarshal(permissionsJSON, &permissions); err != nil {
		return nil, err
	}

	urr.Permissions = make([]domain.Permission, len(permissions))
	for i, p := range permissions {
		urr.Permissions[i] = domain.Permission(p)
	}

	return &urr, nil
}

// SaveUserRoomRole 保存用户房间角色
func (r *UserRoomRoleRepository) SaveUserRoomRole(ctx context.Context, urr *domain.UserRoomRole) error {
	// 序列化权限为 JSONB
	permissionsJSON, err := json.Marshal(urr.Permissions)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO user_room_roles (user_id, room_id, role_name, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, room_id) DO UPDATE SET
			role_name = $3,
			permissions = $4,
			updated_at = $6
	`

	_, err = r.pool.Exec(ctx, sql,
		urr.UserID,
		urr.RoomID,
		urr.RoleName,
		permissionsJSON,
		urr.CreatedAt.Unix(),
		urr.UpdatedAt.Unix(),
	)

	return err
}

// DeleteUserRoomRole 删除用户房间角色
func (r *UserRoomRoleRepository) DeleteUserRoomRole(ctx context.Context, userID, roomID string) error {
	sql := `DELETE FROM user_room_roles WHERE user_id = $1 AND room_id = $2`
	_, err := r.pool.Exec(ctx, sql, userID, roomID)
	return err
}
