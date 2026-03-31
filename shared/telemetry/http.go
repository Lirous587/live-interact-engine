// shared/telemetry/http.go
package telemetry

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// SetupHTTPTracing 为 Gin 引擎设置所有 HTTP 相关的 OTel 集成
// 包括：tracing 中间件 + metrics 中间件
func SetupHTTPTracing(r *gin.Engine, serviceName string) {
	// 1. 官方 OTel Tracing 中间件（自动提取 traceparent）
	r.Use(otelgin.Middleware(serviceName))

	// 2. HTTP Metrics 中间件（请求计数、延迟等）
	r.Use(httpMetricsMiddleware())
}

// httpMetricsMiddleware 记录 HTTP 请求的关键指标
func httpMetricsMiddleware() gin.HandlerFunc {
	// 初始化阶段创建 Meter 和 Instruments
	meter := otel.Meter("http-server")

	requestsTotal, err := meter.Int64Counter(
		"http.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建 http.requests.total 失败: %v", err))
	}

	requestDuration, err := meter.Float64Histogram(
		"http.request.duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建 http.request.duration_seconds 失败: %v", err))
	}

	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		route := ctx.FullPath()

		// 处理空路由
		if route == "" {
			route = "UNKNOWN"

			span := oteltrace.SpanFromContext(ctx.Request.Context())
			if span.IsRecording() {
				span.SetAttributes(attribute.String("http.route", route))
				span.SetName(ctx.Request.Method + " " + route)
			}

		}

		// 记录指标
		duration := time.Since(start).Seconds()
		attrs := metric.WithAttributes(
			attribute.String("http.method", ctx.Request.Method),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", ctx.Writer.Status()),
		)

		requestsTotal.Add(ctx.Request.Context(), 1, attrs)
		requestDuration.Record(ctx.Request.Context(), duration, attrs)
	}
}
