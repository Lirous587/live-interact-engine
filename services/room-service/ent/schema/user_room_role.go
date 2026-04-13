package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserRoomRole holds the schema definition for the UserRoomRole entity.
type UserRoomRole struct {
	ent.Schema
}

// Fields of the UserRoomRole.
func (UserRoomRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}), // 用户 ID（来自 user-service）
		field.UUID("room_id", uuid.UUID{}), // 房间 ID
		field.String("role_name").
			NotEmpty().
			MaxLen(50),
		field.JSON("permissions", []int32{}).
			Default([]int32{}), // 权限列表，JSON 存储以支持灵活扩展
		// created_at: 创建时自动赋值，之后不可修改
		field.Int64("created_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			Immutable(),
		// updated_at: 创建时和每次更新时都自动赋值
		field.Int64("updated_at").
			DefaultFunc(func() int64 { return time.Now().Unix() }).
			UpdateDefault(func() int64 { return time.Now().Unix() }),
	}
}

// Edges of the UserRoomRole.
func (UserRoomRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("room", Room.Type).
			Field("room_id").
			Unique().
			Required(),
	}
}

// Indexes of the UserRoomRole.
func (UserRoomRole) Indexes() []ent.Index {
	return []ent.Index{
		// 复合主键
		index.Fields("user_id", "room_id").Unique(),
		index.Fields("room_id"),
		index.Fields("user_id"),
	}
}
