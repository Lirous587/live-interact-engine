# 分布式链路追踪为什么现在能工作 - 完整解析

## 问题回顾

之前 danmaku-service 的 gRPC span 显示 `Service: api-service` 而不是
`Service: danmaku-service`，链路被"切断"了。

## 根本原因：缺少 TextMapPropagator

### 什么是 TextMapPropagator？

TextMapPropagator 是 OpenTelemetry 中负责**跨进程/跨网络传播链路信息**的组件。它的作用：

1. **发送端**（api-service 调用 gRPC 时）：
   - 从 context 中提取 trace_id、span_id、trace_flags
   - 编码成标准格式（W3C TraceContext）
   - 写入 gRPC metadata 或 HTTP header

2. **接收端**（danmaku-service 收到请求时）：
   - 从 gRPC metadata 或 HTTP header 中读出编码的链路信息
   - 解码还原出 trace_id、span_id
   - **使用同一个 trace_id 创建 server span**

### 没有 Propagator 会怎样？

```
api-service gRPC 调用时
  ↓
context 中有 trace_id = "abc123"
  ↓
但没有 Propagator，无法将 trace_id 写入 gRPC metadata
  ↓
danmaku-service 收到请求
  ↓
读不到 trace_id，生成全新的 trace_id = "xyz789"
  ↓
结果：两个 span 有不同的 trace_id，Jaeger 认为是不同的请求！❌
```

## 解决方案

在 `shared/telemetry/trace.go` 中添加：

```go
// ===== 设置 Propagator（关键！用于链路传播）=====
prop := propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},    // W3C 标准（新）
    propagation.Baggage{},         // 传播业务数据
)
otel.SetTextMapPropagator(prop)
```

## 完整的链路流程

HTTP 请求 → api-service
  ↓
生成/提取 traceparent header（如 traceparent: 00-abc123-def456-01）
  ↓
HTTP Middleware 创建 span，将 trace_id 保存在 context 中
  ↓
调用 gRPC → otelgrpc client handler
  ↓
Client Handler **通过 Propagator** 将 trace_id 写入 gRPC metadata
  ↓
网络传输
  ↓
gRPC 服务端收到请求 → danmaku-service
  ↓
otelgrpc server handler **通过 Propagator** 从 metadata 提取 trace_id
  ↓
使用**同一个 trace_id** 创建 server span
  ↓
Jaeger 看到 trace_id 相同 → 关联成一条链路！

## 为什么之前 Metrics 和 Tracing 混一起也有影响？

两个问题叠加：

1. **缺 Propagator** - 链路信息根本传不过去
2. **Metrics/Tracing 共用 DefaultRegisterer** - 两个服务争用同一个 Prometheus
   registry，可能导致初始化失败

分离后：

- `trace.go` - 纯 Tracing + Propagator → 链路能传播
- `metrics.go` - Metrics 单独初始化 → 不干扰 Tracing

## 关键代码位置

| 文件                                   | 作用                                           |
| -------------------------------------- | ---------------------------------------------- |
| `shared/telemetry/trace.go`            | 初始化 Tracer + **设置 Propagator**            |
| `shared/telemetry/grpc.go`             | gRPC client/server 使用 propagator             |
| `services/api-service/cmd/main.go`     | API 服务启动：`InitTracer()` + `InitMetrics()` |
| `services/danmaku-service/cmd/main.go` | 微服务启动：仅 `InitTracer()`                  |
