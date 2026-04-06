package telemetry

import (
	"context"
	"fmt"
	"live-interact-engine/shared/env"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider 封装所有追踪相关
type TracerProvider struct {
	tp *sdktrace.TracerProvider
}

// InitTracer 初始化追踪
func InitTracer(serviceName string) (*TracerProvider, error) {
	// ===== 配置资源 =====
	res, err := resource.New(context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 resource 失败: %w", err)
	}

	// ===== OTLP HTTP Exporter =====
	endpoint := env.GetString("JAEGER_ENDPOINT", "localhost:4318")
	otlpExp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}

	// ===== Trace Provider =====
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otlpExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(getSampler()),
	)

	// ===== 设置全局 Tracer Provider =====
	otel.SetTracerProvider(tp)

	// ===== 设置 Propagator（关键！用于链路传播）=====
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	log.Printf("Tracer 初始化成功 (service: %s, endpoint: %s)", serviceName, endpoint)

	return &TracerProvider{tp: tp}, nil
}

// Shutdown 优雅关闭 Tracer Provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if err := tp.tp.ForceFlush(ctx); err != nil {
		log.Printf("TracerProvider flush 失败: %v", err)
	}
	if err := tp.tp.Shutdown(ctx); err != nil {
		log.Printf("TracerProvider 关闭失败: %v", err)
	}
	log.Println("TracerProvider 已关闭")
	return nil
}

// GetTracer 获取 tracer
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// 根据环境不同使用不同的采样率
func getSampler() sdktrace.Sampler {
	mode := env.GetString("SERVER_MODE", "dev")
	sampleRate := env.GetFloat64("TRACE_SAMPLE_RATE", 1.0)

	if mode == "dev" || sampleRate >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.TraceIDRatioBased(sampleRate)
}
