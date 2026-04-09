package postgres

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &UserRepository{pool: pool}
}

// GetUser 获取用户
func (r *UserRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password_hash, created_at, updated_at, is_active
		FROM users
		WHERE user_id = $1
	`

	var user domain.User
	var createdAt, updatedAt int64

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
		&updatedAt,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

// SaveUser 保存用户
func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, email, password_hash, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			username = $2,
			email = $3,
			password_hash = $4,
			updated_at = $6,
			is_active = $7
	`

	_, err := r.pool.Exec(ctx, query,
		user.UserID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt.Unix(),
		user.UpdatedAt.Unix(),
		user.IsActive,
	)

	return err
}

// GetUserByEmail 按邮箱获取用户（用于登录）
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT user_id, username, email, password_hash, created_at, updated_at, is_active
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var createdAt, updatedAt int64

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
		&updatedAt,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

// DeleteUser 删除用户
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}