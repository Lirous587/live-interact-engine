package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID {
			return uuid.Must(uuid.NewV7())
		}),
		field.String("username").NotEmpty().MinLen(1).MaxLen(30),
		field.String("email").Unique().NotEmpty(),
		field.String("password_hash").NotEmpty(),
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

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		// 按创建时间排序或过滤
		index.Fields("created_at"),
		// 按用户名查询
		index.Fields("username"),
		// 过滤活跃用户
		index.Fields("is_active"),
		// 复合索引：查询活跃用户并按创建时间排序
		index.Fields("is_active", "created_at"),
	}
}
