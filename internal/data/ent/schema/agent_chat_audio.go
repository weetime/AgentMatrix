package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// AgentChatAudio holds the schema definition for the AgentChatAudio entity.
type AgentChatAudio struct {
	ent.Schema
}

// Fields of the AgentChatAudio.
func (AgentChatAudio) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("主键ID"),
		field.Bytes("audio").
			Comment("音频opus数据"),
	}
}

// Edges of the AgentChatAudio.
func (AgentChatAudio) Edges() []ent.Edge {
	return nil
}

func (AgentChatAudio) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_chat_audio"},
	}
}
