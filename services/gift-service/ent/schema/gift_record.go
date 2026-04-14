package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GiftRecord holds the schema definition for the GiftRecord entity.
type GiftRecord struct {
	ent.Schema
}

// Fields of the GiftRecord.
func (GiftRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("idempotency_key", uuid.UUID{}), // 唯一索引，幂等防重
		field.UUID("user_id", uuid.UUID{}),         // 送礼者 ID
		field.UUID("anchor_id", uuid.UUID{}),       // 主播 ID
		field.UUID("room_id", uuid.UUID{}),
		field.Int64("gift_id").Positive(),
		field.Int64("amount").Positive(),                                               // 送礼金额
		field.Enum("status").Values("pending", "success", "failed").Default("pending"), // 状态机
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the GiftRecord.
func (GiftRecord) Edges() []ent.Edge {
	return nil
}

// Indexes of the GiftRecord.
func (GiftRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("idempotency_key").Unique(), // 防重的唯一索引
		index.Fields("user_id"),                  // 查某用户的送礼记录
		index.Fields("anchor_id"),                // 查主播的收礼记录
		index.Fields("room_id"),                  // 查某房间的送礼记录
		index.Fields("status"),                   // 查待处理流水（pending）
		index.Fields("created_at"),               // 时间范围查询
		index.Fields("status", "created_at"),     // 复合索引：按状态+时间查询（常见查询模式）
	}
}
