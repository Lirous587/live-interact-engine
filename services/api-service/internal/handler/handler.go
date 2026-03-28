package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetUserHandler 示例：获取用户信息
// 演示如何在 handler 中创建子 span，实现分布式链路追踪
func GetUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("user-service")

	// 创建子 span：GetUser 操作
	ctx, span := tracer.Start(ctx, "GetUser",
		trace.WithAttributes(
			attribute.String("user.id", "user123"),
		),
	)
	defer span.End()

	traceID := span.SpanContext().TraceID().String()

	zap.L().Info("GetUser called",
		zap.String("trace_id", traceID),
		zap.String("user_id", "user123"),
	)

	// 模拟数据库查询（另一个子 span）
	dbCtx, dbSpan := tracer.Start(ctx, "db.query")
	dbSpan.SetAttributes(
		attribute.String("db.name", "postgres"),
		attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"),
	)
	time.Sleep(50 * time.Millisecond) // 模拟查询耗时
	dbSpan.End()

	// 模拟缓存检查
	_, cacheSpan := tracer.Start(dbCtx, "cache.get")
	cacheSpan.SetAttributes(
		attribute.String("cache.type", "redis"),
		attribute.String("cache.key", "user:user123"),
	)
	time.Sleep(20 * time.Millisecond) // 模拟缓存查询耗时
	cacheSpan.End()

	c.JSON(http.StatusOK, gin.H{
		"id":       "user123",
		"name":     "张三",
		"email":    "zhangsan@example.com",
		"trace_id": traceID,
	})
}

// CreateTaskHandler 示例：创建任务
// 演示 handler 中的多层级 span，以及对业务关键操作的追踪
func CreateTaskHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("task-service")

	// 创建根 span：CreateTask
	ctx, span := tracer.Start(ctx, "CreateTask",
		trace.WithAttributes(
			attribute.String("user_id", "user123"),
		),
	)
	defer span.End()

	traceID := span.SpanContext().TraceID().String()

	zap.L().Info("CreateTask called", zap.String("trace_id", traceID))

	// 子 span 1: 验证请求体
	ctx, validateSpan := tracer.Start(ctx, "validate.request_body")
	validateSpan.SetAttributes(
		attribute.String("validation.type", "json_schema"),
	)
	time.Sleep(10 * time.Millisecond) // 模拟验证耗时
	validateSpan.End()

	// 子 span 2: 保存数据到数据库
	ctx, dbSpan := tracer.Start(ctx, "db.insert")
	dbSpan.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "tasks"),
	)
	time.Sleep(80 * time.Millisecond) // 模拟 DB 写入耗时
	dbSpan.End()

	// 子 span 3: 发布消息到队列（异步处理）
	ctx, mqSpan := tracer.Start(ctx, "mq.publish")
	mqSpan.SetAttributes(
		attribute.String("mq.broker", "rabbitmq"),
		attribute.String("mq.exchange", "tasks"),
		attribute.String("mq.routing_key", "create.task"),
	)
	time.Sleep(15 * time.Millisecond) // 模拟 MQ 发送耗时
	mqSpan.End()

	c.JSON(http.StatusCreated, gin.H{
		"task_id":  "task-001",
		"status":   "created",
		"trace_id": traceID,
	})
}

// HealthHandler 健康检查端点
func HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("task-service")

	ctx, span := tracer.Start(ctx, "health",
		trace.WithAttributes(
			attribute.String("user_id", "user123"),
		),
	)
	defer span.End()

	traceID := span.SpanContext().TraceID().String()

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"service":  "api-service",
		"trace_id": traceID,
	})
}
