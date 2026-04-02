package telemetry

import (
	"context"
	"fmt"
	"live-interact-engine/shared/env"
	"log"
	"net/http"
	"time"

	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// OTelProviders 保存所有 OTel providers 用于关闭
type OTelProviders struct {
	TracerProvider *tracesdk.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

// InitOTelProviders 初始化所有 OTel providers（包括 metrics 和 tracing）
func InitOTelProviders(serviceName string) (*OTelProviders, error) {
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

	// ===== Tracing Provider (使用 OTLP HTTP) =====
	endpoint := env.GetString("JAEGER_ENDPOINT", "localhost:4318")
	otlpExp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // 开发环境使用 HTTP
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(otlpExp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(getSampler()),
	)
	otel.SetTracerProvider(tp)

	// ===== Metrics Provider =====
	prometheusExp, err := prometheus.New(
		prometheus.WithRegisterer(stdPrometheus.DefaultRegisterer),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Prometheus exporter 失败: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(prometheusExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	log.Printf("OTel providers 初始化成功 (service: %s)", serviceName)

	return &OTelProviders{
		TracerProvider: tp,
		MeterProvider:  mp,
	}, nil
}

// Shutdown 优雅关闭所有 providers
func (op *OTelProviders) Shutdown(ctx context.Context) error {
	// 关闭 TracerProvider（会等待 pending spans 导出）
	if err := op.TracerProvider.ForceFlush(ctx); err != nil {
		log.Printf("TracerProvider flush 失败: %v", err)
	}
	if err := op.TracerProvider.Shutdown(ctx); err != nil {
		log.Printf("TracerProvider 关闭失败: %v", err)
	}

	// 关闭 MeterProvider
	if err := op.MeterProvider.Shutdown(ctx); err != nil {
		log.Printf("MeterProvider 关闭失败: %v", err)
	}

	log.Println("OTel providers 已关闭")
	return nil
}

// GetPrometheusHandler 返回 Prometheus HTTP handler（用于暴露 metrics）
func GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// GetTracer 获取全局 tracer（推荐在各模块中调用）
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// GetMeter 获取全局 meter（推荐在各模块中调用）
func GetMeter(name string) metric.Meter {
	return otel.Meter(name)
}

// StartMetricsServer 在独立端口启动 Prometheus metrics HTTP server
func StartMetricsServer(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", GetPrometheusHandler())

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("启动 Prometheus metrics 服务: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Prometheus metrics 服务启动失败: %v", err)
		}
	}()

	return nil
}

// 根据环境不同使用不同的采样率
func getSampler() tracesdk.Sampler {
	mode := env.GetString("SERVER_MODE", "dev")
	sampleRate := env.GetFloat64("TRACE_SAMPLE_RATE", 1.0) // 可配置

	if mode == "dev" || sampleRate >= 1.0 {
		return tracesdk.AlwaysSample()
	}
	return tracesdk.TraceIDRatioBased(sampleRate)
}
