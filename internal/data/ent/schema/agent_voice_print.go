package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// AgentVoicePrint holds the schema definition for the AgentVoicePrint entity.
type AgentVoicePrint struct {
	ent.Schema
}

// Fields of the AgentVoicePrint.
func (AgentVoicePrint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("声纹ID"),
		field.String("agent_id").
			MaxLen(32).
			Comment("关联的智能体ID"),
		field.String("audio_id").
			MaxLen(32).
			Comment("关联的音频ID"),
		field.String("source_name").
			MaxLen(50).
			Comment("声纹来源的人的姓名"),
		field.String("introduce").
			MaxLen(200).
			Optional().
			Comment("描述声纹来源的这个人"),
		field.Time("create_date").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
		field.Int64("creator").
			Optional().
			Comment("创建者ID"),
		field.Time("update_date").
			Optional().
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("修改时间"),
		field.Int64("updater").
			Optional().
			Comment("修改者ID"),
	}
}

// Edges of the AgentVoicePrint.
func (AgentVoicePrint) Edges() []ent.Edge {
	return nil
}

func (AgentVoicePrint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_voice_print"},
	}
}
