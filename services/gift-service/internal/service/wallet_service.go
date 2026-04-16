package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"live-interact-engine/services/gift-service/internal/domain"
)

type WalletService struct {
	walletRepo   domain.WalletRepository
	walletCache  domain.WalletCache
	walletFilter domain.WalletFilter
}

func NewWalletService(
	walletRepo domain.WalletRepository,
	walletCache domain.WalletCache,
	walletFilter domain.WalletFilter,
) *WalletService {
	return &WalletService{
		walletRepo:   walletRepo,
		walletCache:  walletCache,
		walletFilter: walletFilter,
	}
}

// GetWallet 获取钱包信息（先查缓存，缓存未命中则查数据库）
func (s *WalletService) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	// 先从缓存获取余额
	cachedBalance, err := s.walletCache.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 如果缓存有值，从数据库获取完整钱包信息，但用缓存中的余额
	wallet, err := s.walletRepo.GetWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, nil
	}

	// 用缓存中的余额覆盖数据库的余额（因为Redis是实时的）
	wallet.Balance = cachedBalance
	return wallet, nil
}

// DeductBalance 使用Redis Lua原子扣款
func (s *WalletService) DeductBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	// 检查过滤器，如果未初始化则从 DB 加载
	if !s.walletFilter.Exists(ctx, userID) {
		wallet, err := s.walletRepo.GetWallet(ctx, userID)
		if err != nil {
			return 0, err
		}

		balance := int64(0)
		if wallet != nil {
			balance = wallet.Balance
		}

		// 写入 Redis 缓存
		if err := s.walletCache.SetBalance(ctx, userID, balance); err != nil {
			return 0, err
		}

		// 标记为已初始化
		if err := s.walletFilter.Add(ctx, userID); err != nil {
			zap.L().Error("add wallet filter failed", zap.String("user_id", userID.String()), zap.Error(err))
		}
	}

	return s.walletCache.DeductByLua(ctx, userID, amount, idempotencyKey)
}

// IncrementBalance 使用Redis Lua原子增加余额
func (s *WalletService) IncrementBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	// 检查过滤器，如果未初始化则从 DB 加载
	if !s.walletFilter.Exists(ctx, userID) {
		wallet, err := s.walletRepo.GetWallet(ctx, userID)
		if err != nil {
			return 0, err
		}

		balance := int64(0)
		if wallet != nil {
			balance = wallet.Balance
		}

		// 写入 Redis 缓存
		if err := s.walletCache.SetBalance(ctx, userID, balance); err != nil {
			return 0, err
		}

		// 标记为已初始化
		if err := s.walletFilter.Add(ctx, userID); err != nil {
			zap.L().Error("add wallet filter failed", zap.String("user_id", userID.String()), zap.Error(err))
		}
	}

	return s.walletCache.IncrementByLua(ctx, userID, amount, idempotencyKey)
}

// InitializeWallet 初始化新用户的钱包（从 user-service 调用）
func (s *WalletService) InitializeWallet(ctx context.Context, userID uuid.UUID) error {
	// 1. 创建钱包记录到数据库（初始余额为 0）
	wallet := &domain.Wallet{
		UserID:  userID,
		Balance: 0,
	}

	if err := s.walletRepo.SaveWallet(ctx, wallet); err != nil {
		return err
	}

	// 2. 写入 Redis 缓存
	if err := s.walletCache.SetBalance(ctx, userID, 0); err != nil {
		zap.L().Warn("failed to set wallet balance to cache",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// 不返回错误，因为 DB 已经保存了
	}

	// 3. 标记为已初始化
	if err := s.walletFilter.Add(ctx, userID); err != nil {
		zap.L().Warn("failed to add wallet filter",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// 不返回错误
	}

	return nil
}
