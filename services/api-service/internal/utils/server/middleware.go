package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// SetupMiddlewares 设置所有中间件（Tracing + Logging + Metrics）
func SetupMiddlewares(r *gin.Engine, serviceName string) {
	// 1. 官方 OTel Tracing 中间件
	// 自动处理：HTTP Header traceparent 提取、请求属性记录、panic 捕获等
	r.Use(otelgin.Middleware(serviceName))

	// 2. 错误处理中间件 → 记录异常堆栈
	r.Use(errorHandler())

	// 3. 优化的指标中间件 → Prometheus
	r.Use(MetricsMiddleware())
}

func printBusinessStack(err error) {
	// 获取完整错误栈
	stackTrace := fmt.Sprintf("%+v", err)
	lines := strings.Split(stackTrace, "\n")

	// 错误消息
	if len(lines) > 0 {
		log.Printf("\n\n")
		log.Printf("业务逻辑错误: %s\n", lines[0])
	}

	// 记录已打印的栈帧数量
	framePrinted := 0
	maxBusinessFrames := 3 // 最多打印栈帧条数

	// 逐行检查并不做任何修改，保持原始格式
	for i := 0; i < len(lines)-1 && framePrinted < maxBusinessFrames; i++ {
		currentLine := lines[i]
		nextLine := lines[i+1]

		// 只检查是否为业务相关行，但完全保持原始格式
		if strings.Contains(currentLine, "internal") &&
			!strings.Contains(currentLine, "github.com/gin-gonic") &&
			!strings.Contains(currentLine, "net/http") &&
			!strings.Contains(currentLine, "net/http") &&
			!strings.Contains(currentLine, "internal/common/server") &&
			strings.Contains(nextLine, ".go:") {
			log.Println(currentLine)
			log.Println(nextLine)
			framePrinted++
		}
	}

	// 如果还有更多栈帧但已达到限制
	totalBusinessFrames := countBusinessFrames(lines)
	if framePrinted == maxBusinessFrames && framePrinted < totalBusinessFrames {
		log.Printf("一共%d条栈帧,实际打印%d条 (更多栈帧已省略)\n", totalBusinessFrames, maxBusinessFrames)
	}
}

// 计算业务栈帧总数
func countBusinessFrames(lines []string) int {
	count := 0
	for i := 0; i < len(lines)-1; i++ {
		currentLine := lines[i]
		nextLine := lines[i+1]

		if strings.Contains(currentLine, "internal") &&
			!strings.Contains(currentLine, "reskit") &&
			!strings.Contains(currentLine, "github.com/gin-gonic") &&
			!strings.Contains(currentLine, "net/http") &&
			strings.Contains(nextLine, ".go:") {
			count++
		}
	}
	return count
}

// 错误链追踪 用于开发环境
func errorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		// 处理错误
		if len(ctx.Errors) > 0 {
			for _, e := range ctx.Errors {
				// 记录详细错误日志
				//log.Printf("Error: %+v\n", e.Err)

				// 使用自定义格式化错误栈
				printBusinessStack(e.Err)
			}
		}
	}
}

// MetricsMiddleware 优化的指标中间件
// Meter 和 Instruments 在闭包外部初始化，确保每次请求只是轻量级的 Add/Record 操作
func MetricsMiddleware() gin.HandlerFunc {
	// 在初始化阶段就创建 Meter 和 Counter/Histogram，而不是在每次请求时创建
	meter := otel.Meter("gin-http-server")

	// HTTP 请求总数计数器
	requestsTotal, err := meter.Int64Counter(
		"http.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建 requestsTotal counter 失败: %v", err))
	}

	// HTTP 请求耗时直方图
	requestDuration, err := meter.Float64Histogram(
		"http.request.duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建 requestDuration histogram 失败: %v", err))
	}

	// 返回实际处理请求的中间件函数
	return func(ctx *gin.Context) {
		start := time.Now()

		// 交给下一个中间件/路由处理请求
		ctx.Next()

		// 请求处理完毕后，记录指标
		duration := time.Since(start).Seconds()
		status := ctx.Writer.Status()
		method := ctx.Request.Method

		// 使用 FullPath 而不是 URL.Path
		// 例如：/api/user/123 会统一记录为 /api/user/:id
		// 这样可以避免路径参数导致的"指标爆炸"（cardinality explosion）
		path := ctx.FullPath()
		if path == "" {
			path = "UNKNOWN"
		}

		// 构建公共属性
		attrs := metric.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", path),
			attribute.Int("http.status_code", status),
		)

		// 记录指标
		// 直接使用外部已经初始化好的 requestsTotal 和 requestDuration
		requestsTotal.Add(ctx.Request.Context(), 1, attrs)
		requestDuration.Record(ctx.Request.Context(), duration, attrs)
	}
}
