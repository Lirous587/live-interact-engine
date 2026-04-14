package service

import (
	"context"

	"github.com/google/uuid"

	"live-interact-engine/services/gift-service/internal/domain"
)

type WalletService struct {
	walletRepo  domain.WalletRepository
	walletCache domain.WalletCache
}

func NewWalletService(
	walletRepo domain.WalletRepository,
	walletCache domain.WalletCache,
) *WalletService {
	return &WalletService{
		walletRepo:  walletRepo,
		walletCache: walletCache,
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

// SaveWallet 保存钱包（同时更新数据库和缓存）
func (s *WalletService) SaveWallet(ctx context.Context, wallet *domain.Wallet) error {
	// 同时更新数据库和缓存
	if err := s.walletRepo.SaveWallet(ctx, wallet); err != nil {
		return err
	}

	// 将余额同步到缓存
	if err := s.walletCache.SetBalance(ctx, wallet.UserID, wallet.Balance); err != nil {
		// 缓存更新失败不影响主流程
		_ = err
	}

	return nil
}

// DeductBalance 使用Redis Lua原子扣款
func (s *WalletService) DeductBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	return s.walletCache.DeductByLua(ctx, userID, amount, idempotencyKey)
}

// IncrementBalance 使用Redis Lua原子增加余额
func (s *WalletService) IncrementBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	return s.walletCache.IncrementByLua(ctx, userID, amount, idempotencyKey)
}
