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

func (Wallet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(), // 用户 ID 唯一（一用户一钱包）
		index.Fields("updated_at"),       // 按更新时间查询（余额变动审计）
	}
}
