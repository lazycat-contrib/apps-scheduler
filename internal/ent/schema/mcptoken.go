package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// MCPToken holds metadata for a user-managed MCP access token.
type MCPToken struct {
	ent.Schema
}

// Fields of the MCPToken.
func (MCPToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("token_hash").NotEmpty().Unique().Sensitive(),
		field.String("token_prefix").NotEmpty(),
		field.String("user_id").NotEmpty(),
		field.String("user_role").Default("USER"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}

// Edges of the MCPToken.
func (MCPToken) Edges() []ent.Edge {
	return nil
}
