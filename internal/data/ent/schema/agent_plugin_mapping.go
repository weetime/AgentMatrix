package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AgentPluginMapping holds the schema definition for the AgentPluginMapping entity.
type AgentPluginMapping struct {
	ent.Schema
}

// Fields of the AgentPluginMapping.
func (AgentPluginMapping) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("主键"),
		field.String("agent_id").
			MaxLen(32).
			Comment("智能体ID"),
		field.String("plugin_id").
			MaxLen(32).
			Comment("插件ID"),
		field.String("param_info").
			SchemaType(map[string]string{
				dialect.MySQL:    "json",
				dialect.Postgres: "jsonb",
			}).
			Comment("插件参数(JSON格式)"),
	}
}

// Edges of the AgentPluginMapping.
func (AgentPluginMapping) Edges() []ent.Edge {
	return nil
}

// Indexes of the AgentPluginMapping.
func (AgentPluginMapping) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id", "plugin_id").
			Unique().
			StorageKey("uk_agent_provider"),
	}
}

func (AgentPluginMapping) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_plugin_mapping"},
	}
}
