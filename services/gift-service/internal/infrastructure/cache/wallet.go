package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"live-interact-engine/services/gift-service/internal/domain"
)

// walletCache Redis 钱包缓存实现
type walletCache struct {
	redis              *redis.Client
	prefix             string // 例如 "wallet:"
	luaDeductScript    string
	luaIncrementScript string
}

// NewWalletCache 创建钱包缓存，返回实现 domain.WalletCache 接口的实例
func NewWalletCache(redis *redis.Client) domain.WalletCache {
	// Lua 脚本的设计思想：
	// 1. 检查幂等性（防止重复扣款）
	// 2. 检查余额
	// 3. 扣款 + 记录幂等性
	// 4. 全部原子执行
	luaDeductScript := `
		local key = KEYS[1]
		local idempotencyKey = KEYS[2]
		local amount = tonumber(ARGV[1])

		-- 检查幂等性：如果这个 idempotency_key 已经处理过，直接返回失败
		if redis.call('EXISTS', idempotencyKey) == 1 then
			return {0, redis.call('GET', key) or 0}  -- 幂等防重，不再处理
		end

		-- 获取当前余额
		local balance = tonumber(redis.call('GET', key) or 0)

		-- 检查余额是否足够
		if balance < amount then
			return {-1, balance}  -- -1 表示余额不足
		end

		-- 原子扣款
		balance = balance - amount
		redis.call('SET', key, balance)

		-- 记录幂等性（设置过期时间24小时，防止 hash 表无限增长）
		redis.call('SETEX', idempotencyKey, 86400, '1')

		return {1, balance}  -- 1 表示扣款成功，返回新余额
	`

	luaIncrementScript := `
		local key = KEYS[1]
		local idempotencyKey = KEYS[2]
		local amount = tonumber(ARGV[1])

		-- 检查幂等性
		if redis.call('EXISTS', idempotencyKey) == 1 then
			return {0, redis.call('GET', key) or 0}  -- 已处理过
		end

		-- 原子增加余额
		local balance = tonumber(redis.call('GET', key) or 0)
		balance = balance + amount
		redis.call('SET', key, balance)

		-- 记录幂等性
		redis.call('SETEX', idempotencyKey, 86400, '1')

		return {1, balance}  -- 1 表示成功
	`

	return &walletCache{
		redis:              redis,
		prefix:             "wallet:",
		luaDeductScript:    luaDeductScript,
		luaIncrementScript: luaIncrementScript,
	}
}

// GetBalance 获取用户余额
func (c *walletCache) GetBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	key := c.prefix + userID.String()
	balance, err := c.redis.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 用户未初始化，返回 0
		}
		return 0, err
	}
	return balance, nil
}

// SetBalance 设置用户余额
func (c *walletCache) SetBalance(ctx context.Context, userID uuid.UUID, balance int64) error {
	key := c.prefix + userID.String()
	return c.redis.Set(ctx, key, balance, 0).Err()
}

// DeductByLua 使用 Lua 脚本原子扣款
func (c *walletCache) DeductByLua(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	key := c.prefix + userID.String()
	idempotencyKeyStr := "idempotency:" + idempotencyKey.String()

	// 执行 Lua 脚本
	result, err := c.redis.Eval(ctx, c.luaDeductScript, []string{key, idempotencyKeyStr}, amount).Result()
	if err != nil {
		return 0, err
	}

	// 解析返回值
	results := result.([]interface{})
	code := int64(results[0].(int64))
	newBalance := int64(results[1].(int64))

	switch code {
	case 1:
		return newBalance, nil // 扣款成功
	case 0:
		return newBalance, domain.ErrInvalidAmount // 幂等防重，已处理过
	case -1:
		return newBalance, domain.ErrInsufficientBalance // 余额不足
	default:
		return 0, fmt.Errorf("unknown error code: %d", code)
	}
}

// IncrementByLua 使用 Lua 脚本原子增加余额
func (c *walletCache) IncrementByLua(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey uuid.UUID) (int64, error) {
	key := c.prefix + userID.String()
	idempotencyKeyStr := "idempotency:" + idempotencyKey.String()

	result, err := c.redis.Eval(ctx, c.luaIncrementScript, []string{key, idempotencyKeyStr}, amount).Result()
	if err != nil {
		return 0, err
	}

	results := result.([]interface{})
	code := int64(results[0].(int64))
	newBalance := int64(results[1].(int64))

	switch code {
	case 1:
		return newBalance, nil // 增加成功
	case 0:
		return newBalance, domain.ErrInvalidAmount // 幂等防重
	default:
		return 0, fmt.Errorf("unknown error code: %d", code)
	}
}

// DeleteBalance 删除余额缓存（测试用）
func (c *walletCache) DeleteBalance(ctx context.Context, userID uuid.UUID) error {
	key := c.prefix + userID.String()
	return c.redis.Del(ctx, key).Err()
}
