package schema

import "entgo.io/ent"

// Gift holds the schema definition for the Gift entity.
type Gift struct {
	ent.Schema
}

// Fields of the Gift.
func (Gift) Fields() []ent.Field {
	return nil
}

// Edges of the Gift.
func (Gift) Edges() []ent.Edge {
	return nil
}
