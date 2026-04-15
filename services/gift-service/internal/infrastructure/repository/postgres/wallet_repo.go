package postgres

import (
	"context"

	"live-interact-engine/services/gift-service/ent"
	"live-interact-engine/services/gift-service/internal/domain"

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

// SaveWallet 保存或更新钱包
// SaveWallet 保存或更新钱包
func (r *WalletRepository) SaveWallet(ctx context.Context, wallet *domain.Wallet) error {
	err := r.client.Wallet.
		Create().
		SetUserID(wallet.UserID).
		SetBalance(wallet.Balance).
		SetVersionNumber(wallet.VersionNumber).
		OnConflictColumns(entwallet.FieldUserID).
		UpdateNewValues().
		Exec(ctx)

	return err
}

// GetWallet 根据用户ID获取钱包
func (r *WalletRepository) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	entWallet, err := r.client.Wallet.Query().
		Where(entwallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
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

func entWalletToDomain(entWallet *ent.Wallet) *domain.Wallet {
	return &domain.Wallet{
		UserID:        entWallet.UserID,
		Balance:       entWallet.Balance,
		VersionNumber: entWallet.VersionNumber,
		CreatedAt:     entWallet.CreatedAt,
		UpdatedAt:     entWallet.UpdatedAt,
	}
}
