package postgres

import (
	"context"

	"live-interact-engine/services/gift-service/ent"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/pkg/types"

	entwallet "live-interact-engine/services/gift-service/ent/wallet"

	"github.com/google/uuid"
)

type WalletRepository struct {
	client *ent.Client
}

func NewWalletRepository(client *ent.Client) domain.WalletRepository {
	return &WalletRepository{
		client: client,
	}
}

// CreateWallet 创建新钱包
func (r *WalletRepository) CreateWallet(ctx context.Context, wallet *domain.Wallet) error {
	err := r.client.Wallet.
		Create().
		SetUserID(wallet.UserID).
		SetBalance(wallet.Balance).
		SetVersionNumber(wallet.VersionNumber).
		Exec(ctx)

	return err
}

// UpdateWallet 更新钱包余额（使用乐观锁，仅 Consumer 调用）
func (r *WalletRepository) UpdateWallet(ctx context.Context, wallet *domain.Wallet) error {
	// 基于版本号的乐观锁更新
	result, err := r.client.Wallet.
		Update().
		Where(
			entwallet.UserIDEQ(wallet.UserID),
			entwallet.VersionNumberEQ(wallet.VersionNumber),
		).
		SetBalance(wallet.Balance).
		SetVersionNumber(wallet.VersionNumber + 1).
		Save(ctx)

	if err != nil {
		return err
	}

	// 如果没有更新任何行，说明版本号不匹配（被其他进程更新了）
	if result == 0 {
		return types.ErrVersionConflict
	}

	return nil
}

// GetWallet 根据用户ID获取钱包
func (r *WalletRepository) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	entWallet, err := r.client.Wallet.Query().
		Where(entwallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, types.ErrWalletNotFound
		}
		return nil, err
	}
	return entWalletToDomain(entWallet), nil
}

// DeleteWallet 删除钱包
func (r *WalletRepository) DeleteWallet(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.Wallet.Delete().
		Where(entwallet.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// Tx 开启一个新事务
func (r *WalletRepository) Tx(ctx context.Context) (domain.Tx, error) {
	return r.client.Tx(ctx)
}

// UpdateWalletTx 在事务内更新钱包余额（使用乐观锁）
func (r *WalletRepository) UpdateWalletTx(ctx context.Context, tx domain.Tx, wallet *domain.Wallet) error {
	entTx := tx.(*ent.Tx)
	result, err := entTx.Wallet.
		Update().
		Where(
			entwallet.UserIDEQ(wallet.UserID),
			entwallet.VersionNumberEQ(wallet.VersionNumber),
		).
		SetBalance(wallet.Balance).
		SetVersionNumber(wallet.VersionNumber + 1).
		Save(ctx)

	if err != nil {
		return err
	}

	if result == 0 {
		return types.ErrVersionConflict
	}

	return nil
}

func entWalletToDomain(entWallet *ent.Wallet) *domain.Wallet {
	return &domain.Wallet{
		UserID:        entWallet.UserID,
		Balance:       entWallet.Balance,
		VersionNumber: entWallet.VersionNumber,
		CreatedAt:     entWallet.CreatedAt,
		UpdatedAt:     entWallet.UpdatedAt,
	}
}
