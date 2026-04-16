package events

import (
	"context"
	"encoding/json"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"live-interact-engine/services/gift-service/internal/domain"
)

type Consumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	giftRecordRepo domain.GiftRecordRepository
	walletRepo     domain.WalletRepository
	walletService  domain.WalletService
}

// NewConsumer 创建消费者
func NewConsumer(
	rabbitmqURL string,
	giftRecordRepo domain.GiftRecordRepository,
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
		conn:           conn,
		channel:        ch,
		giftRecordRepo: giftRecordRepo,
		walletRepo:     walletRepo,
		walletService:  walletService,
	}, nil
}

// Start 启动消费者（阻塞）
func (c *Consumer) Start(ctx context.Context) error {
	// 从队列获取消息
	msgs, err := c.channel.Consume(
		QueueName, // 队列名
		"",        // consumer tag（空表示自动生成）
		false,     // auto-ack（必须 false，手动 ack）
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		zap.L().Error("failed to consume from queue", zap.Error(err))
		return err
	}

	zap.L().Info("gift consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				zap.L().Info("message channel closed")
				return nil
			}

			// 处理消息
			c.handleMessage(ctx, &msg)
		}
	}
}

// handleMessage 处理单条消息
func (c *Consumer) handleMessage(ctx context.Context, msg *amqp.Delivery) {
	// 解析事件
	var event GiftSendSuccessEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		zap.L().Error("failed to unmarshal event",
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		// 解析失败，NACK（会重新入队，最终进死信队列）
		msg.Nack(false, false)
		return
	}

	// 1. 插入 GiftRecord（可能有唯一约束冲突，正常）
	giftRecord := &domain.GiftRecord{
		IdempotencyKey: event.IdempotencyKey,
		UserID:         event.UserID,
		AnchorID:       event.AnchorID,
		RoomID:         event.RoomID,
		GiftID:         event.GiftID,
		Amount:         event.Amount,
		Status:         domain.GiftRecordStatusSuccess,
	}

	err = c.giftRecordRepo.SaveGiftRecord(ctx, giftRecord)
	if err != nil {
		// 检查是否是唯一约束冲突（表示已处理过）
		if c.isUniqueConstraintError(err) {
			zap.L().Warn("gift record already exists (idempotent)",
				zap.String("idempotency_key", event.IdempotencyKey.String()),
				zap.String("user_id", event.UserID.String()))
			// 已处理过，ACK（不再处理）
			msg.Ack(false)
			return
		}

		// 其他错误，NACK（重新入队）
		zap.L().Error("failed to save gift record",
			zap.Error(err),
			zap.String("idempotency_key", event.IdempotencyKey.String()))
		msg.Nack(false, true) // requeue=true，重新入队
		return
	}

	// 2. 更新 Wallet 余额（主播的余额增加）
	wallet, err := c.walletRepo.GetWallet(ctx, event.AnchorID)
	if err != nil {
		zap.L().Error("failed to get anchor wallet",
			zap.Error(err),
			zap.String("anchor_id", event.AnchorID.String()))
		msg.Nack(false, true)
		return
	}

	if wallet == nil {
		// 主播钱包不存在，创建一个
		wallet = &domain.Wallet{
			UserID:        event.AnchorID,
			Balance:       event.Amount,
			VersionNumber: 1,
		}
	} else {
		wallet.Balance += event.Amount
		wallet.VersionNumber++
	}

	err = c.walletRepo.SaveWallet(ctx, wallet)
	if err != nil {
		zap.L().Error("failed to save anchor wallet",
			zap.Error(err),
			zap.String("anchor_id", event.AnchorID.String()))
		msg.Nack(false, true)
		return
	}

	// 3. 更新缓存（Redis + Filter）
	_, err = c.walletService.IncrementBalance(ctx, event.AnchorID, event.Amount, event.IdempotencyKey)
	if err != nil {
		zap.L().Warn("failed to increment cache balance (non-critical)",
			zap.Error(err),
			zap.String("anchor_id", event.AnchorID.String()))
		// 缓存更新失败不影响，因为有 DB 作为真实来源
		// 继续 ACK
	}

	// 4. ACK（表示成功处理）
	msg.Ack(false)

	zap.L().Info("gift event processed successfully",
		zap.String("user_id", event.UserID.String()),
		zap.String("anchor_id", event.AnchorID.String()),
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
