package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Wallet holds the schema definition for the Wallet entity.
type Wallet struct {
	ent.Schema
}

// Fields of the Wallet.
func (Wallet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Int64("balance").Default(0),
		field.Int64("version_number").Default(0), // 乐观锁版本号 用于CAS
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Wallet.
func (Wallet) Edges() []ent.Edge {
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
