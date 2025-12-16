package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AgentContextProvider holds the schema definition for the AgentContextProvider entity.
type AgentContextProvider struct {
	ent.Schema
}

// Fields of the AgentContextProvider.
func (AgentContextProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("主键"),
		field.String("agent_id").
			MaxLen(32).
			Comment("智能体ID"),
		field.String("context_providers").
			SchemaType(map[string]string{
				dialect.MySQL:    "json",
				dialect.Postgres: "jsonb",
			}).
			Optional().
			Comment("上下文源配置"),
		field.Int64("creator").
			Optional().
			Comment("创建者"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
		field.Int64("updater").
			Optional().
			Comment("更新者"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("更新时间"),
	}
}

// Edges of the AgentContextProvider.
func (AgentContextProvider) Edges() []ent.Edge {
	return nil
}

// Indexes of the AgentContextProvider.
func (AgentContextProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id").
			StorageKey("idx_agent_id"),
	}
}

func (AgentContextProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_context_provider"},
	}
}
