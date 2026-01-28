package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// NotifyConfig holds the schema definition for the NotifyConfig entity.
type NotifyConfig struct {
	ent.Schema
}

// Fields of the NotifyConfig.
func (NotifyConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique().Immutable(),
		field.String("user_id").Unique().NotEmpty(),
		field.String("send_key").Default(""),
		field.Bool("enabled").Default(false),
		field.Bool("on_success").Default(true),
		field.Bool("on_failure").Default(true),
	}
}

// Edges of the NotifyConfig.
func (NotifyConfig) Edges() []ent.Edge {
	return nil
}
