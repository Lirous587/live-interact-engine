package schema

import "entgo.io/ent"

// Room holds the schema definition for the Room entity.
type Room struct {
	ent.Schema
}

// Fields of the Room.
func (Room) Fields() []ent.Field {
	return nil
}

// Edges of the Room.
func (Room) Edges() []ent.Edge {
	return nil
}
