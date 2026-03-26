# live-interact-engine

> 一个具备实时弹幕广播与强一致性礼物结算的高并发直播后端引擎。
>
> 致力于研究读写扩散和分布式下的一致性问题。

## 核心价值

- **高并发支撑**: 单机可支持数万 WebSocket 长连接的弹幕推送
- **强一致性保障**: 利用 Redis Lua 脚本实现原子操作，无资金丢失风险
- **架构可靠性**: 微服务解耦，模块独立扩展，单点故障隔离
- **研究意义**: 研究分布式系统中的读写扩散、一致性保证等前沿问题

---

## 微服务架构

本项目包含三个核心微服务，采用 gRPC 通信，支持独立部署与扩展：

### 1. **api-service** (API 网关)

- 系统的统一入口点
- HTTP 请求处理与路由聚合
- 用户鉴权与流量控制
- 负责调用后端 gRPC 服务

### 2. **gift-service** (礼物与打赏)

- 核心业务逻辑：打赏、扣减、对账
- 余额原子扣减（Redis Lua）
- 强一致性保证
- 交易流水的持久化

### 3. **danmaku-service** (弹幕推送)

- WebSocket 长连接管理
- 房间与连接池管理
- 时间窗口聚合策略（高并发场景）
- 实时消息广播

---

## 技术栈

| 层次          | 技术选型              | 说明               |
| ------------- | --------------------- | ------------------ |
| **语言**      | Go 1.22+              | 高性能并发编程     |
| **数据库**    | PostgreSQL (pgx)      | 交易数据持久化     |
| **缓存**      | Redis (go-redis)      | 热点数据与原子操作 |
| **消息队列**  | RabbitMQ (amqp091-go) | 服务间异步通信     |
| **WebSocket** | gorilla/websocket     | 长连接实时推送     |
| **RPC**       | gRPC                  | 服务间同步通信     |

---

## 项目结构

```
live-interact-engine/
├── services/
│   ├── api-service/          # API 网关服务
│   ├── gift-service/         # 打赏服务
│   └── danmaku-service/      # 弹幕服务
├── shared/                   # 共享库与工具
├── proto/                    # gRPC 协议定义
├── infra/                    # 基础设施配置
│   ├── development/
│   └── production/
├── tools/                    # 辅助工具
├── build/                    # 构建产物
├── docs/                     # 文档
├── go.mod                    # Go 模块定义
├── Makefile                  # 构建脚本
├── Dockerfile                # 容器配置
└── README.md
```

---


## 核心设计原则

### 架构设计

1. **微服务独立**: 三个服务各司其职，通过 gRPC 通信
2. **面向接口编程**: 所有外部依赖通过接口抽象，便于测试 Mock
3. **依赖注入**: 显式传递依赖，避免隐式初始化
4. **标准 Go Layout**: 遵循项目目录标准

### 编码规范

1. **错误处理**: 必须使用 `fmt.Errorf("...: %w", err)` 包装错误上下文
2. **并发安全**: Goroutine 必须捕获 panic，严格处理 Context 传递
3. **日志输出**: 结构化日志，记录关键业务事件
4. **代码注释**: 中文核心逻辑注释，关键流程补充说明

---

## 测试

```bash
# 单元测试
make test

# 集成测试
make test-integration

# 基准测试
make bench-gift-service
make bench-danmaku-service
```

---

## 文档

- [项目规划详览](./docs/PLANNING.md) - 详细的项目规划与任务清单
- [API 文档](./docs/API.md) - gRPC 接口定义
- [部署指南](./docs/DEPLOYMENT.md) - 生产环境部署
- [开发指南](./docs/DEVELOPMENT.md) - 开发流程与规范

---

## 贡献指南

欢迎提交 Issue 与 PR！请确保：

1. 代码符合项目编码规范
2. 新增功能需附带单元测试
3. 提交说明清晰明了

---

## 许可证

MIT License
