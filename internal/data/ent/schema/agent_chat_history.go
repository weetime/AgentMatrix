package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// AgentChatHistory holds the schema definition for the AgentChatHistory entity.
type AgentChatHistory struct {
	ent.Schema
}

// Fields of the AgentChatHistory.
func (AgentChatHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.String("mac_address").
			MaxLen(50).
			Optional().
			Comment("MAC地址"),
		field.String("agent_id").
			MaxLen(32).
			Optional().
			Comment("智能体ID"),
		field.String("session_id").
			MaxLen(50).
			Optional().
			Comment("会话ID"),
		field.Int8("chat_type").
			Comment("消息类型: 1-用户, 2-智能体"),
		field.String("content").
			MaxLen(1024).
			Optional().
			Comment("聊天内容"),
		field.String("audio_id").
			MaxLen(32).
			Optional().
			Comment("音频ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime(3)",
				dialect.Postgres: "timestamp(3)",
			}).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime(3)",
				dialect.Postgres: "timestamp(3)",
			}).
			Comment("更新时间"),
	}
}

// Edges of the AgentChatHistory.
func (AgentChatHistory) Edges() []ent.Edge {
	return nil
}

func (AgentChatHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_chat_history"},
	}
}
