package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// WalletTransaction holds the schema definition for the WalletTransaction entity.
type WalletTransaction struct {
	ent.Schema
}

// Fields of the WalletTransaction.
func (WalletTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("idempotency_key", uuid.UUID{}).Unique().Immutable(),

		field.Int("type"), // 0=送礼, 1=充值

		field.UUID("payer_id", uuid.UUID{}),                       // 谁付钱
		field.UUID("payee_id", uuid.UUID{}).Optional().Nillable(), // 谁收钱（可为空）

		field.Int64("amount"),
		field.UUID("room_id", uuid.UUID{}).Optional().Nillable(), // 仅送礼
		field.UUID("gift_id", uuid.UUID{}).Optional().Nillable(), // 仅送礼

		field.String("status").Default("success"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the WalletTransaction.
func (WalletTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("payer_id", "type"),
		index.Fields("payee_id", "type"),
		index.Fields("idempotency_key").Unique(),
	}
}

// Edges of the WalletTransaction.
func (WalletTransaction) Edges() []ent.Edge {
	return nil
}
