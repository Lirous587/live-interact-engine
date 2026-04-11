package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(41), // 使用UUID作为ID
		field.String("username").NotEmpty(),
		field.String("email").Unique().NotEmpty(),
		field.String("password_hash").NotEmpty(),
		field.Int64("created_at").Default(0),
		field.Int64("updated_at").Default(0),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("created_at"),
	}
}
