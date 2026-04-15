# 礼物数据初始化

## 概述

本目录提供两种方式初始化礼物数据：

1. **SQL脚本** (`02-init-gifts.sql`) - 直接数据库脚本
2. **Go程序** (`cmd/seed/main.go`) - Go语言初始化程序

## 礼物数据说明

### 礼物分类

#### 普通礼物 (5个)

- **点赞** (1金币) - 最基础的礼物
- **玫瑰花** (10金币)
- **钻戒** (100金币)
- **劳斯莱斯** (1000金币)
- **城堡** (5000金币)

#### VIP专属礼物 (2个)

- **VIP勋章** (50金币) - 仅VIP用户可送
- **皇冠** (200金币) - 仅VIP用户可送

#### 限时礼物 (2个)

- **新春红包** (88金币) - 新年限定
- **圣诞礼物** (66金币) - 圣诞限定

#### 下线礼物 (1个)

- **过期礼物** (1金币) - 展示已下线状态

## 使用方法

### 方式1: 使用SQL脚本（推荐用于数据库初始化）

```bash
# PostgreSQL命令行
psql -U postgres -h localhost -d gift_service < infra/postgres/02-init-gifts.sql

# 或使用docker exec
docker exec -i <postgres_container> psql -U postgres -d gift_service < 02-init-gifts.sql
```

### 方式2: 使用Go程序（推荐用于测试）

#### 最简单：使用命令行参数指定数据库

```bash
# 在项目根目录执行
cd services/gift-service

# 指定自定义数据库连接
go run ./cmd/seed/main.go -dsn "postgres://user:password@localhost:5432/gift_service?sslmode=disable"
```

#### 或使用环境变量

**PowerShell:**

```powershell
$env:DATABASE_DSN = "postgres://user:password@localhost:5432/gift_service?sslmode=disable"
go run ./cmd/seed/main.go
```

**CMD/Batch:**

```batch
set DATABASE_DSN=postgres://user:password@localhost:5432/gift_service?sslmode=disable
go run ./cmd/seed/main.go
```

**Bash/Linux:**

```bash
export DATABASE_DSN="postgres://user:password@localhost:5432/gift_service?sslmode=disable"
go run ./cmd/seed/main.go
```

#### Docker内部运行（使用默认配置）

```bash
# 在Docker container内运行
go run ./cmd/seed/main.go
```

## 环境变量配置

Go程序会读取以下环境变量：

- `DATABASE_DSN` - 数据库连接字符串（默认从env.go读取）
- `REDIS_ADDR` - Redis地址（用于缓存）

## 礼物数据结构

```go
type Gift struct {
    ID          uuid.UUID      // 礼物ID
    Name        string         // 礼物名称
    Description string         // 礼物描述
    IconURL     string         // 图标URL
    CacheKey    string         // Redis缓存键
    Price       int64          // 价格（单位：金币）
    VIPOnly     bool           // 是否VIP专属
    Status      GiftStatus     // 状态 (online/offline/limited_time)
    CreatedAt   time.Time      // 创建时间
    UpdatedAt   time.Time      // 更新时间
}

// GiftStatus 枚举
type GiftStatus string

const (
    GiftStatusOnline      GiftStatus = "online"        // 在线销售
    GiftStatusOffline     GiftStatus = "offline"       // 已下线
    GiftStatusLimitedTime GiftStatus = "limited_time"  // 限时销售
)
```

## 测试验证

初始化完成后，可以通过以下方式验证：

### 1. 检查数据库

```sql
SELECT id, name, price, vip_only, status FROM gifts ORDER BY price;
```

### 2. 调用API

```bash
# 获取所有在线礼物
curl -X GET "http://localhost:8080/v1/gift/list" \
  -H "Authorization: Bearer <token>"

# 返回示例：
{
  "code": 0,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "点赞",
      "description": "最简单的表达方式",
      "icon_url": "https://example.com/thumbs-up.png",
      "price": 1,
      "vip_only": false,
      "status": "online"
    },
    ...
  ]
}
```

## 自定义礼物

如需添加更多礼物，可以：

1. **修改SQL脚本**: 在 `02-init-gifts.sql`
   中 INSERT 语句添加新的VALUE行
2. **修改Go程序**: 在 `cmd/seed/main.go` 的 `gifts`
   数组中添加新的GiftSeed

示例：

```go
{
    ID:          uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440099")),
    Name:        "自定义礼物",
    Description: "我的特殊礼物",
    IconURL:     "https://example.com/custom.png",
    CacheKey:    "gift:custom",
    Price:       999,
    VIPOnly:     false,
    Status:      domain.GiftStatusOnline,
},
```

## 注意事项

1. **ID唯一性** - 确保每个礼物的UUID不重复
2. **CacheKey** - 建议使用 `gift:<name>` 格式，便于Redis缓存管理
3. **状态一致性** - VIP专属礼物应设置 `VIPOnly: true`
4. **价格验证** - 所有礼物价格必须大于0

## 常见问题

### Q: 运行Go程序时报错"database not found"

A: 确保PostgreSQL数据库已创建，检查 `01-init-databases.sql` 是否已执行

### Q: 数据库中已有礼物，再次初始化是否会冲突？

A: SQL脚本默认会插入，如需更新可启用 `ON CONFLICT`
子句；Go程序会直接覆盖

### Q: 如何删除所有礼物重新初始化？

A: 执行以下SQL：

```sql
DELETE FROM gifts;
```
