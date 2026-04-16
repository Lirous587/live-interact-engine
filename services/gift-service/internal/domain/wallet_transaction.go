package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WalletTransactionType 钱包交易类型
type WalletTransactionType int

const (
	WalletTransactionTypeGiftSend WalletTransactionType = iota // 0
	WalletTransactionTypeRecharge                              // 1
)

// WalletTransaction 钱包交易记录
type WalletTransaction struct {
	IdempotencyKey uuid.UUID
	Type           WalletTransactionType
	PayerID        uuid.UUID
	PayeeID        uuid.UUID // 可为空
	Amount         int64
	RoomID         uuid.UUID // 仅送礼
	GiftID         uuid.UUID // 仅送礼
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// WalletTransactionRepository 钱包交易记录仓储
type WalletTransactionRepository interface {
	// SaveWalletTransactionTx 在事务内保存交易记录
	SaveWalletTransactionTx(ctx context.Context, tx Tx, transaction *WalletTransaction) error

	// GetWalletTransaction 根据幂等性键获取交易记录
	GetWalletTransaction(ctx context.Context, idempotencyKey uuid.UUID) (*WalletTransaction, error)

	// ListTransactionsByPayer 获取付款人的交易记录
	ListTransactionsByPayer(ctx context.Context, payerID uuid.UUID, offset, limit int) ([]*WalletTransaction, error)

	// ListTransactionsByPayee 获取收款人的交易记录
	ListTransactionsByPayee(ctx context.Context, payeeID uuid.UUID, offset, limit int) ([]*WalletTransaction, error)
}
