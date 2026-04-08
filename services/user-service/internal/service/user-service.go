package service

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
)

type UserServiceImpl struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) (domain.UserService, error) {
	return &UserServiceImpl{
		userRepo: userRepo,
	}, nil
}

// 获取用户基本信息
func (s *UserServiceImpl) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetUser(ctx, userID)
}
