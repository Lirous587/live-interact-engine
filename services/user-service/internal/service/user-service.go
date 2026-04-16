package service

import (
	"context"
	"time"

	"live-interact-engine/services/user-service/internal/domain"
	"live-interact-engine/services/user-service/pkg/types"
	"live-interact-engine/shared/crypto"
	pb "live-interact-engine/shared/proto/gift"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo            domain.UserRepository
	passwordMgr         *crypto.PasswordManager
	tokenService        domain.TokenService
	giftWalletClient    pb.WalletServiceClient
}

func NewUserService(userRepo domain.UserRepository, giftWalletClient pb.WalletServiceClient) (domain.UserService, error) {
	return &UserService{
		userRepo:            userRepo,
		passwordMgr:         crypto.NewPasswordManager(12),
		giftWalletClient:    giftWalletClient,
	}, nil
}

// SetTokenService setter 注入 tokenService（解决循环依赖）
func (s *UserService) SetTokenService(tokenService domain.TokenService) {
	s.tokenService = tokenService
}

// 获取用户基本信息
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, types.ErrUserNotFound
	}
	return user, nil
}

// Register 注册新用户 (不返回 token，客户端需要调用 Login 获取 token)
func (s *UserService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	// 1. 验证输入
	if username == "" || email == "" || password == "" {
		return nil, types.ErrInvalidInput
	}

	// 2. 检查邮箱是否已注册
	existing, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, types.ErrEmailAlreadyRegistered
	}

	// 3. 创建新用户
	now := time.Now()
	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	passwordHash, err := s.passwordMgr.HashPassword(password)
	if err != nil {
		return nil, types.ErrPasswordHashFailed
	}

	user := &domain.User{
		UserID:       userID,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	// 4. 保存用户到数据库
	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	// 5. 初始化钱包（调用 gift-service gRPC）
	initWalletReq := &pb.InitializeWalletRequest{
		UserId: userID.String(),
	}
	_, err = s.giftWalletClient.InitializeWallet(ctx, initWalletReq)
	if err != nil {
		// 钱包初始化失败，记录错误但不阻止用户注册
		zap.L().Error("failed to initialize wallet for user",
			zap.String("user_id", userID.String()),
			zap.Error(err))
	}

	return user, nil
}

// Login 登录
func (s *UserService) Login(ctx context.Context, email, password string, metadata domain.UserIdentityMetadata) (*domain.User, *domain.TokenPair, error) {
	// 1. 验证输入
	if email == "" || password == "" {
		return nil, nil, types.ErrInvalidInput
	}

	// 2. 从数据库查询用户
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, types.ErrInvalidCredentials
	}

	// 3. 检查用户是否激活
	if !user.IsActive {
		return nil, nil, types.ErrUserInactive
	}

	// 4. 验证密码
	if !s.passwordMgr.VerifyPassword(user.PasswordHash, password) {
		return nil, nil, types.ErrInvalidCredentials
	}

	// 5. 生成 token pair
	now := time.Now()
	identity := &domain.UserIdentity{
		UserID:               user.UserID,
		UserIdentityMetadata: metadata,
	}

	tokenPair, err := s.tokenService.GenTokenPair(ctx, &domain.TokenPayload{
		Identity:  identity,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(5 * time.Minute).Unix(),
	})
	if err != nil {
		return nil, nil, err
	}

	return user, tokenPair, nil
}
