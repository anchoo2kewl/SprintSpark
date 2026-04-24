package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Milestone holds the schema definition for the Milestone entity.
type Milestone struct {
	ent.Schema
}

// Fields of the Milestone.
func (Milestone) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("project_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional().Nillable(),
		field.String("color").Default("#5e6ad2"),
		field.Time("target_date").Optional().Nillable(),
		field.String("status").Default("active"),
		field.Int("sort_order").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Milestone.
func (Milestone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("milestones").Unique().Required().Field("project_id"),
		edge.To("tasks", Task.Type),
	}
}

// Indexes of the Milestone.
func (Milestone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("status"),
	}
}
