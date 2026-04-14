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

// Indexes of the Wallet.
func (Wallet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(), // 用户 ID 唯一（一用户一钱包）
		index.Fields("updated_at"),       // 按更新时间查询（余额变动审计）
	}
}
