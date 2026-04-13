package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Room holds the schema definition for the Room entity.
type Room struct {
	ent.Schema
}

// Fields of the Room.
func (Room) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID {
			return uuid.Must(uuid.NewV7())
		}),
		field.UUID("owner_id", uuid.UUID{}), // 房间所有者 ID（来自 user-service）
		field.String("title").NotEmpty().MinLen(1).MaxLen(30),
		field.Text("description").Optional(),
		// created_at: 创建时自动赋值，之后不可修改
		field.Int64("created_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			Immutable(),
		// updated_at: 创建时和每次更新时都自动赋值
		field.Int64("updated_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			UpdateDefault(func() int64 { return time.Now().Unix() }),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the Room.
func (Room) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_room_roles", UserRoomRole.Type).
			Ref("room"),
	}
}

// Indexes of the Room.
func (Room) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id"),
		index.Fields("is_active"),
		index.Fields("is_active", "created_at"),
	}
}
