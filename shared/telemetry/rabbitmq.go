package telemetry

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// amqpHeadersCarrier 将 AMQP headers 适配为 OTel TextMapCarrier。
type amqpHeadersCarrier amqp.Table

func (c amqpHeadersCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		s, ok := v.(string)
		if ok {
			return s
		}
	}
	return ""
}

func (c amqpHeadersCarrier) Set(key, value string) {
	c[key] = value
}

func (c amqpHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// TracedPublish 包装 RabbitMQ publish，并自动注入 trace context 到 headers。
// publishFn 只需要执行底层 publish 动作。
func TracedPublish(
	ctx context.Context,
	exchange string,
	routingKey string,
	msg amqp.Publishing,
	publishFn func(context.Context, string, string, amqp.Publishing) error,
) error {
	tracer := otel.GetTracerProvider().Tracer("rabbitmq")

	ctx, span := tracer.Start(ctx, "rabbitmq.publish")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.destination", exchange),
		attribute.String("messaging.rabbitmq.routing_key", routingKey),
	)

	// 通用 JSON 字段增强
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err == nil {
		if v, ok := body["user_id"].(string); ok && v != "" {
			span.SetAttributes(attribute.String("messaging.user_id", v))
		}
		if v, ok := body["room_id"].(string); ok && v != "" {
			span.SetAttributes(attribute.String("messaging.room_id", v))
		}
	}

	if msg.Headers == nil {
		msg.Headers = make(amqp.Table)
	}
	carrier := amqpHeadersCarrier(msg.Headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	msg.Headers = amqp.Table(carrier)

	if err := publishFn(ctx, exchange, routingKey, msg); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

// TracedConsume 包装 RabbitMQ consumer handler，并从 headers 提取 trace context。
func TracedConsume(
	delivery amqp.Delivery,
	handler func(context.Context, amqp.Delivery) error,
) error {
	carrier := amqpHeadersCarrier(delivery.Headers)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	tracer := otel.GetTracerProvider().Tracer("rabbitmq")
	ctx, span := tracer.Start(ctx, "rabbitmq.consume")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.operation", "consume"),
		attribute.String("messaging.destination", delivery.Exchange),
		attribute.String("messaging.rabbitmq.routing_key", delivery.RoutingKey),
	)

	var body map[string]any
	if err := json.Unmarshal(delivery.Body, &body); err == nil {
		if v, ok := body["user_id"].(string); ok && v != "" {
			span.SetAttributes(attribute.String("messaging.user_id", v))
		}
		if v, ok := body["room_id"].(string); ok && v != "" {
			span.SetAttributes(attribute.String("messaging.room_id", v))
		}
	}

	if err := handler(ctx, delivery); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}
