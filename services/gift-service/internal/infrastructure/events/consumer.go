package events

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"live-interact-engine/services/gift-service/internal/domain"
)

type Consumer struct {
	conn                   *amqp.Connection
	channel                *amqp.Channel
	walletTransactionRepo  domain.WalletTransactionRepository
	walletRepo             domain.WalletRepository
	walletService          domain.WalletService
}

// NewConsumer 创建消费者
func NewConsumer(
	rabbitmqURL string,
	walletTransactionRepo domain.WalletTransactionRepository,
	walletRepo domain.WalletRepository,
	walletService domain.WalletService,
) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		zap.L().Error("failed to connect to RabbitMQ", zap.Error(err))
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		zap.L().Error("failed to open channel", zap.Error(err))
		conn.Close()
		return nil, err
	}

	// 设置 QoS：每个消费者最多处理 1 条消息（保证顺序处理）
	err = ch.Qos(1, 0, false)
	if err != nil {
		zap.L().Error("failed to set qos", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:                  conn,
		channel:               ch,
		walletTransactionRepo: walletTransactionRepo,
		walletRepo:            walletRepo,
		walletService:         walletService,
	}, nil
}

// Start 启动消费者（阻塞）
func (c *Consumer) Start(ctx context.Context) error {
	// 1. 声明并绑定礼物队列
	_, err := c.channel.QueueDeclare(
		QueueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "gift-dlx",
		},
	)
	if err != nil {
		zap.L().Error("failed to declare gift queue", zap.Error(err))
		return err
	}

	// 绑定礼物队列到Exchange（仅订阅 gift.send.success 消息）
	err = c.channel.QueueBind(
		QueueName,
		"gift.send.success",
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind gift queue", zap.Error(err))
		return err
	}

	// 2. 消费礼物发送事件
	giftMsgs, err := c.channel.Consume(
		QueueName, // 队列名
		"",        // consumer tag（空表示自动生成）
		false,     // auto-ack（必须 false，手动 ack）
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		zap.L().Error("failed to consume from gift queue", zap.Error(err))
		return err
	}

	zap.L().Info("gift consumer started, waiting for messages...")

	// 声明充值队列
	_, err = c.channel.QueueDeclare(
		"wallet.recharge.queue",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		zap.L().Error("failed to declare wallet recharge queue", zap.Error(err))
		return err
	}

	// 绑定充值队列到Exchange
	err = c.channel.QueueBind(
		"wallet.recharge.queue",
		"wallet.recharge",
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind wallet recharge queue", zap.Error(err))
		return err
	}

	// 消费充值事件
	rechargeMsgs, err := c.channel.Consume(
		"wallet.recharge.queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to consume from wallet recharge queue", zap.Error(err))
		return err
	}

	zap.L().Info("wallet recharge consumer started, waiting for messages...")

	// 并发处理两个队列的消息
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-giftMsgs:
			if !ok {
				zap.L().Info("gift message channel closed")
				return nil
			}
			c.handleGiftSendSuccessEvent(ctx, &msg)

		case msg, ok := <-rechargeMsgs:
			if !ok {
				zap.L().Info("wallet recharge message channel closed")
				return nil
			}
			c.handleWalletRechargeEvent(ctx, &msg)
		}
	}
}

// handleGiftSendSuccessEvent 处理礼物发送成功事件
func (c *Consumer) handleGiftSendSuccessEvent(ctx context.Context, msg *amqp.Delivery) {
	var event GiftSendSuccessEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		zap.L().Error("failed to unmarshal event",
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		msg.Nack(false, false)
		return
	}

	// 获取钱包信息
	wallet, err := c.walletRepo.GetWallet(ctx, event.AnchorID)
	if err != nil {
		zap.L().Error("failed to get wallet", zap.Error(err))
		msg.Nack(false, true)
		return
	}

	transaction := &domain.WalletTransaction{
		IdempotencyKey: event.IdempotencyKey,
		Type:           domain.WalletTransactionTypeGiftSend,
		PayerID:        event.UserID,
		PayeeID:        event.AnchorID,
		Amount:         event.Amount,
		RoomID:         event.RoomID,
		GiftID:         event.GiftID,
		Status:         "success",
	}

	// 事务内执行：保存交易记录 + 更新钱包
	updated := false
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, err := c.walletRepo.Tx(ctx)
		if err != nil {
			zap.L().Error("failed to begin transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 保存交易记录
		err = c.walletTransactionRepo.SaveWalletTransactionTx(ctx, tx, transaction)
		if err != nil {
			tx.Rollback()
			if c.isUniqueConstraintError(err) {
				// 已处理过，ACK
				msg.Ack(false)
				zap.L().Warn("wallet transaction already exists (idempotent)",
					zap.String("idempotency_key", event.IdempotencyKey.String()))
				return
			}
			zap.L().Error("failed to save wallet transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 更新钱包余额
		wallet.Balance += event.Amount
		wallet.VersionNumber++
		err = c.walletRepo.UpdateWalletTx(ctx, tx, wallet)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "Version conflict") {
				if attempt < maxAttempts-1 {
					delayMs := 50 * (1 << uint(attempt))
					if delayMs > 5000 {
						delayMs = 5000
					}
					zap.L().Warn("version conflict, retrying",
						zap.String("anchor_id", event.AnchorID.String()),
						zap.Int("attempt", attempt+1))
					time.Sleep(time.Duration(delayMs) * time.Millisecond)
				}
				continue
			}
			zap.L().Error("failed to update wallet", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			zap.L().Error("failed to commit transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}
		updated = true
		break
	}

	if !updated {
		zap.L().Error("failed to update wallet after retries")
		msg.Nack(false, false)
		return
	}

	// 更新缓存
	c.walletService.IncrementBalance(ctx, event.AnchorID, event.Amount, event.IdempotencyKey)

	msg.Ack(false)
	zap.L().Info("gift event processed successfully",
		zap.String("anchor_id", event.AnchorID.String()),
		zap.Int64("amount", event.Amount))
}

// handleWalletRechargeEvent 处理钱包充值事件
func (c *Consumer) handleWalletRechargeEvent(ctx context.Context, msg *amqp.Delivery) {
	var event WalletRechargeEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		zap.L().Error("failed to unmarshal wallet recharge event",
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		msg.Nack(false, false)
		return
	}

	// 获取钱包信息
	wallet, err := c.walletRepo.GetWallet(ctx, event.UserID)
	if err != nil {
		zap.L().Error("failed to get wallet", zap.Error(err))
		msg.Nack(false, true)
		return
	}

	transaction := &domain.WalletTransaction{
		IdempotencyKey: event.IdempotencyKey,
		Type:           domain.WalletTransactionTypeRecharge,
		PayerID:        event.UserID,
		PayeeID:        uuid.Nil,
		Amount:         event.Amount,
		Status:         "success",
	}

	// 事务内执行：保存交易记录 + 更新钱包
	updated := false
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, err := c.walletRepo.Tx(ctx)
		if err != nil {
			zap.L().Error("failed to begin transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 保存交易记录
		err = c.walletTransactionRepo.SaveWalletTransactionTx(ctx, tx, transaction)
		if err != nil {
			tx.Rollback()
			if c.isUniqueConstraintError(err) {
				// 已处理过，ACK
				msg.Ack(false)
				zap.L().Warn("wallet transaction already exists (idempotent)",
					zap.String("idempotency_key", event.IdempotencyKey.String()))
				return
			}
			zap.L().Error("failed to save wallet transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 更新钱包余额
		wallet.Balance += event.Amount
		wallet.VersionNumber++
		err = c.walletRepo.UpdateWalletTx(ctx, tx, wallet)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "Version conflict") {
				if attempt < maxAttempts-1 {
					delayMs := 50 * (1 << uint(attempt))
					if delayMs > 5000 {
						delayMs = 5000
					}
					zap.L().Warn("version conflict, retrying",
						zap.String("user_id", event.UserID.String()),
						zap.Int("attempt", attempt+1))
					time.Sleep(time.Duration(delayMs) * time.Millisecond)
				}
				continue
			}
			zap.L().Error("failed to update wallet", zap.Error(err))
			msg.Nack(false, true)
			return
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			zap.L().Error("failed to commit transaction", zap.Error(err))
			msg.Nack(false, true)
			return
		}
		updated = true
		break
	}

	if !updated {
		zap.L().Error("failed to update wallet after retries")
		msg.Nack(false, false)
		return
	}

	msg.Ack(false)
	zap.L().Info("wallet recharge event processed successfully",
		zap.String("user_id", event.UserID.String()),
		zap.Int64("amount", event.Amount))
}

// isUniqueConstraintError 检查是否是唯一约束错误
func (c *Consumer) isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 唯一约束错误信息通常包含 "unique constraint"
	errMsg := err.Error()
	return strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "duplicate key")
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
