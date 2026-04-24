package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TaskDependency holds the schema definition for the TaskDependency entity (junction table).
type TaskDependency struct {
	ent.Schema
}

// Fields of the TaskDependency.
func (TaskDependency) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("task_id"),
		field.Int64("depends_on_id"),
		field.String("dependency_type").Default("blocks"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the TaskDependency.
func (TaskDependency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).Ref("dependencies").Unique().Required().Field("task_id"),
		edge.From("depends_on", Task.Type).Ref("dependents").Unique().Required().Field("depends_on_id"),
	}
}

// Indexes of the TaskDependency.
func (TaskDependency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("depends_on_id"),
		index.Fields("task_id", "depends_on_id").Unique(),
	}
}
