package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tx 事务接口
type Tx interface {
	Commit() error
	Rollback() error
}

// Wallet 用户钱包领域模型
type Wallet struct {
	UserID        uuid.UUID
	Balance       int64
	VersionNumber int64 // 版本号（乐观锁用于数据库层）
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// HasSufficientBalance 检查余额是否足够
func (w *Wallet) HasSufficientBalance(amount int64) bool {
	return w.Balance >= amount
}

// Deduct 扣款（仅业务逻辑验证，实际扣款在 Redis Lua 中执行）
func (w *Wallet) Deduct(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !w.HasSufficientBalance(amount) {
		return ErrInsufficientBalance
	}
	w.Balance -= amount
	w.VersionNumber++
	return nil
}

// Recharge 充值
func (w *Wallet) Recharge(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	w.Balance += amount
	w.VersionNumber++
	return nil
}

type WalletService interface {
	GetWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error)
	DeductBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error)
	IncrementBalance(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error)
	InitializeWallet(ctx context.Context, userID uuid.UUID) error
}

type WalletRepository interface {
	// CreateWallet 创建新钱包（用户注册时）
	CreateWallet(ctx context.Context, wallet *Wallet) error

	// UpdateWallet 更新钱包余额（Consumer 调用，带乐观锁）
	// 返回 ErrVersionConflict 如果版本不符
	UpdateWallet(ctx context.Context, wallet *Wallet) error

	GetWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	DeleteWallet(ctx context.Context, userID uuid.UUID) error

	// Tx 开启事务
	Tx(ctx context.Context) (Tx, error)

	// UpdateWalletTx 在事务内更新钱包（带乐观锁）
	UpdateWalletTx(ctx context.Context, tx Tx, wallet *Wallet) error
}

// WalletCache 钱包缓存接口
type WalletCache interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (int64, error)

	SetBalance(ctx context.Context, userID uuid.UUID, balance int64) error

	// DeductByLua 使用 Lua 脚本原子扣款
	// 返回：新余额, 错误
	// 错误包括：ErrInsufficientBalance（余额不足）
	DeductByLua(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error)

	// IncrementByLua 使用 Lua 脚本原子增加余额（充值）
	IncrementByLua(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error)
}

type WalletFilter interface {
	// Exists 判断用户钱包是否已初始化
	Exists(ctx context.Context, userID uuid.UUID) bool

	// Add 添加用户（标记为已初始化）
	Add(ctx context.Context, userID uuid.UUID) error
}
