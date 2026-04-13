package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Mute struct {
	ent.Schema
}

func (Mute) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID {
			return uuid.Must(uuid.NewV7())
		}),
		field.UUID("room_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("admin_id", uuid.UUID{}),
		field.String("reason").Optional(),
		field.Int64("duration"),
		field.Int64("muted_at").DefaultFunc(func() int64 { return time.Now().Unix() }),
		field.Int64("expires_at").DefaultFunc(func() int64 { return time.Now().Unix() }),
		field.Int64("created_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			Immutable(),
		field.Int64("updated_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			UpdateDefault(func() int64 { return time.Now().Unix() }),
	}
}

func (Mute) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id", "user_id").Unique(),
		index.Fields("expires_at"),
		index.Fields("room_id", "expires_at"),
	}
}
