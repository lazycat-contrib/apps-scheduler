package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Schedule holds the schema definition for the Schedule entity.
type Schedule struct {
	ent.Schema
}

// Fields of the Schedule.
func (Schedule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("app_id").NotEmpty(),
		field.String("app_title").Default(""),
		field.Enum("operation").Values("resume", "pause").Default("pause"),
		field.JSON("week_days", []int{}).Default([]int{}),
		field.Int("hour").Min(0).Max(23).Default(0),
		field.Int("minute").Min(0).Max(59).Default(0),
		field.Bool("enabled").Default(true),
		field.String("creator").Immutable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Schedule.
func (Schedule) Edges() []ent.Edge {
	return nil
}
