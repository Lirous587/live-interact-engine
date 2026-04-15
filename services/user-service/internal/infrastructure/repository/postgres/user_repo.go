package postgres

import (
	"context"
	"live-interact-engine/services/user-service/ent"
	entuser "live-interact-engine/services/user-service/ent/user"
	"live-interact-engine/services/user-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) domain.UserRepository {
	return &UserRepository{client: client}
}

// GetUser 获取用户
func (r *UserRepository) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	entUser, err := r.client.User.Query().
		Where(entuser.IDEQ(userID)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		UserID:       entUser.ID,
		Username:     entUser.Username,
		Email:        entUser.Email,
		PasswordHash: entUser.PasswordHash,
		CreatedAt:    time.Unix(entUser.CreatedAt, 0),
		UpdatedAt:    time.Unix(entUser.UpdatedAt, 0),
		IsActive:     entUser.IsActive,
	}, nil
}

// SaveUser 保存用户（插入或更新）
func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	err := r.client.User.Create().
		SetID(user.UserID).
		SetUsername(user.Username).
		SetEmail(user.Email).
		SetPasswordHash(user.PasswordHash).
		SetCreatedAt(user.CreatedAt.Unix()).
		SetUpdatedAt(user.UpdatedAt.Unix()).
		SetIsActive(user.IsActive).
		OnConflictColumns(entuser.FieldID).
		UpdateNewValues().
		Exec(ctx)

	return err
}

// GetUserByEmail 按邮箱获取用户（用于登录）
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	entUser, err := r.client.User.Query().
		Where(entuser.EmailEQ(email)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		UserID:       entUser.ID,
		Username:     entUser.Username,
		Email:        entUser.Email,
		PasswordHash: entUser.PasswordHash,
		CreatedAt:    time.Unix(entUser.CreatedAt, 0),
		UpdatedAt:    time.Unix(entUser.UpdatedAt, 0),
		IsActive:     entUser.IsActive,
	}, nil
}

// DeleteUser 删除用户
func (r *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := r.client.User.
		DeleteOneID(userID).
		Exec(ctx)

	// 如果用户不存在，Ent会返回错误，但这不是我们关心的错误
	// 因为删除不存在的用户应该被视为成功
	if ent.IsNotFound(err) {
		return nil
	}

	return err
}
