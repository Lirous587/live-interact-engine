package events

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	ExchangeName = "gift-events"
	QueueName    = "gift-send-success"
	RoutingKey   = "gift.send.success"

	DeadLetterExchange = "gift-dlx"
	DeadLetterQueue    = "gift-dlq"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher 创建发布者
func NewPublisher(rabbitmqURL string) (*Publisher, error) {
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

	// 声明死信 Exchange
	err = ch.ExchangeDeclare(
		DeadLetterExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare dlx exchange", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	// 声明死信队列
	_, err = ch.QueueDeclare(
		DeadLetterQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare dlq", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	// 绑定死信队列
	err = ch.QueueBind(
		DeadLetterQueue,
		RoutingKey,
		DeadLetterExchange,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind dlq", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	// 声明主 Exchange (Fanout - 为了后续扩展)
	err = ch.ExchangeDeclare(
		ExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare exchange", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	// 声明主队列（带死信配置）
	_, err = ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange": DeadLetterExchange,
		},
	)
	if err != nil {
		zap.L().Error("failed to declare queue", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	// 绑定队列到 Exchange
	err = ch.QueueBind(
		QueueName,
		RoutingKey,
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind queue", zap.Error(err))
		ch.Close()
		conn.Close()
		return nil, err
	}

	zap.L().Info("RabbitMQ publisher initialized")

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// PublishGiftSendSuccess 发布礼物成功事件
func (p *Publisher) PublishGiftSendSuccess(ctx context.Context, event *GiftSendSuccessEvent) error {
	if event == nil {
		return nil
	}

	// 添加时间戳
	event.Timestamp = time.Now().Unix()

	// Marshal 为 JSON
	body, err := json.Marshal(event)
	if err != nil {
		zap.L().Error("failed to marshal event", zap.Error(err))
		return err
	}

	// 发送到 RabbitMQ（3次重试）
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		err = p.channel.PublishWithContext(
			ctx,
			ExchangeName,
			RoutingKey,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
			},
		)
		if err == nil {
			zap.L().Info("gift event published",
				zap.String("user_id", event.UserID.String()),
				zap.String("anchor_id", event.AnchorID.String()),
				zap.Int64("amount", event.Amount),
			)
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(100*(attempt+1)) * time.Millisecond)
	}

	zap.L().Error("failed to publish event after retries",
		zap.Error(lastErr),
		zap.String("user_id", event.UserID.String()),
	)
	return lastErr
}

// Close 关闭连接
func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
