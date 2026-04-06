package telemetry

import (
	"fmt"
	"log"
	"net/http"
	"time"

	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitMetrics 初始化 Prometheus Metrics（仅 API 服务使用，微服务无需）
func InitMetrics(serviceName string) error {
	// ===== 配置资源 =====
	res, err := resource.New(nil,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return fmt.Errorf("创建 resource 失败: %w", err)
	}

	// ===== Metrics Provider =====
	prometheusExp, err := prometheus.New(
		prometheus.WithRegisterer(stdPrometheus.DefaultRegisterer),
	)
	if err != nil {
		return fmt.Errorf("创建 Prometheus exporter 失败: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(prometheusExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	log.Printf("Metrics 初始化成功 (service: %s)", serviceName)
	return nil
}

// StartMetricsServer 在独立端口启动 Prometheus metrics HTTP server
func StartMetricsServer(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

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
