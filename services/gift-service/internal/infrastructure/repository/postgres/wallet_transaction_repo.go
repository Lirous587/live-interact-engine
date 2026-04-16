package postgres

import (
	"context"

	"live-interact-engine/services/gift-service/ent"
	"live-interact-engine/services/gift-service/internal/domain"

	entwt "live-interact-engine/services/gift-service/ent/wallettransaction"

	"github.com/google/uuid"
)

type WalletTransactionRepository struct {
	client *ent.Client
}

func NewWalletTransactionRepository(client *ent.Client) domain.WalletTransactionRepository {
	return &WalletTransactionRepository{
		client: client,
	}
}

// SaveWalletTransactionTx 在事务内保存交易记录（幂等性防重）
func (r *WalletTransactionRepository) SaveWalletTransactionTx(ctx context.Context, tx domain.Tx, transaction *domain.WalletTransaction) error {
	entTx := tx.(*ent.Tx)
	err := entTx.WalletTransaction.Create().
		SetIdempotencyKey(transaction.IdempotencyKey).
		SetType(int(transaction.Type)).
		SetPayerID(transaction.PayerID).
		SetNillablePayeeID(nilUUIDToNil(transaction.PayeeID)).
		SetAmount(transaction.Amount).
		SetNillableRoomID(nilUUIDToNil(transaction.RoomID)).
		SetNillableGiftID(nilUUIDToNil(transaction.GiftID)).
		SetStatus(transaction.Status).
		OnConflictColumns(entwt.FieldIdempotencyKey).
		UpdateNewValues().
		Exec(ctx)

	return err
}

// GetWalletTransaction 根据幂等性键获取交易记录
func (r *WalletTransactionRepository) GetWalletTransaction(ctx context.Context, idempotencyKey uuid.UUID) (*domain.WalletTransaction, error) {
	entTx, err := r.client.WalletTransaction.Query().
		Where(entwt.IdempotencyKeyEQ(idempotencyKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entWalletTransactionToDomain(entTx), nil
}

// ListTransactionsByPayer 获取付款人的交易记录
func (r *WalletTransactionRepository) ListTransactionsByPayer(ctx context.Context, payerID uuid.UUID, offset, limit int) ([]*domain.WalletTransaction, error) {
	entTxs, err := r.client.WalletTransaction.Query().
		Where(entwt.PayerIDEQ(payerID)).
		Order(entwt.ByCreatedAt()).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.WalletTransaction, len(entTxs))
	for i, tx := range entTxs {
		transactions[i] = entWalletTransactionToDomain(tx)
	}
	return transactions, nil
}

// ListTransactionsByPayee 获取收款人的交易记录
func (r *WalletTransactionRepository) ListTransactionsByPayee(ctx context.Context, payeeID uuid.UUID, offset, limit int) ([]*domain.WalletTransaction, error) {
	entTxs, err := r.client.WalletTransaction.Query().
		Where(entwt.PayeeIDEQ(payeeID)).
		Order(entwt.ByCreatedAt()).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.WalletTransaction, len(entTxs))
	for i, tx := range entTxs {
		transactions[i] = entWalletTransactionToDomain(tx)
	}
	return transactions, nil
}

func entWalletTransactionToDomain(entTx *ent.WalletTransaction) *domain.WalletTransaction {
	return &domain.WalletTransaction{
		IdempotencyKey: entTx.IdempotencyKey,
		Type:           domain.WalletTransactionType(entTx.Type),
		PayerID:        entTx.PayerID,
		PayeeID:        pointerUUIDToValue(entTx.PayeeID),
		Amount:         entTx.Amount,
		RoomID:         pointerUUIDToValue(entTx.RoomID),
		GiftID:         pointerUUIDToValue(entTx.GiftID),
		Status:         entTx.Status,
		CreatedAt:      entTx.CreatedAt,
		UpdatedAt:      entTx.UpdatedAt,
	}
}

// nilUUIDToNil 将零值UUID转换为nil
func nilUUIDToNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

// pointerUUIDToValue 将指针UUID转换为值，如果为nil则返回uuid.Nil
func pointerUUIDToValue(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}
