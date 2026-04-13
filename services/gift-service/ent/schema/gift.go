package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Gift holds the schema definition for the Gift entity.
type Gift struct {
	ent.Schema
}

// Fields of the Gift.
func (Gift) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MinLen(1).MaxLen(50),
		field.String("description").Optional().MaxLen(200),
		field.String("icon_url").Optional().MaxLen(500),
		field.String("cache_key").Immutable(), // cache_key
		field.Int64("price").Positive(),       // 礼物价格（平台币），必须 > 0
		field.Bool("vip_only").Default(false), // 仅会员可以赠送
		field.String("special_effect").Optional().MaxLen(50),
		field.Enum("status").Values("online", "offline", "limited_time").Default("online"), // 状态管理
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Gift.
func (Gift) Edges() []ent.Edge {
	return nil
}

func (Gift) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cache_key").Unique(), // cache_key 唯一（加快 Redis 预热时的查询）
		index.Fields("status"),             // 按状态查询礼物（online/offline）
		index.Fields("price"),              // 按价格查询礼物
		index.Fields("created_at"),         // 按时间排序
	}
}
