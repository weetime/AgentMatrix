package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// AgentTemplate holds the schema definition for the AgentTemplate entity.
type AgentTemplate struct {
	ent.Schema
}

// Fields of the AgentTemplate.
func (AgentTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("智能体唯一标识"),
		field.String("agent_code").
			MaxLen(36).
			Optional().
			Comment("智能体编码"),
		field.String("agent_name").
			MaxLen(64).
			Optional().
			Comment("智能体名称"),
		field.String("asr_model_id").
			MaxLen(32).
			Optional().
			Comment("语音识别模型标识"),
		field.String("vad_model_id").
			MaxLen(64).
			Optional().
			Comment("语音活动检测标识"),
		field.String("llm_model_id").
			MaxLen(32).
			Optional().
			Comment("大语言模型标识"),
		field.String("vllm_model_id").
			MaxLen(32).
			Optional().
			Comment("VLLM模型标识"),
		field.String("tts_model_id").
			MaxLen(32).
			Optional().
			Comment("语音合成模型标识"),
		field.String("tts_voice_id").
			MaxLen(32).
			Optional().
			Comment("音色标识"),
		field.String("mem_model_id").
			MaxLen(32).
			Optional().
			Comment("记忆模型标识"),
		field.String("intent_model_id").
			MaxLen(32).
			Optional().
			Comment("意图模型标识"),
		field.Int32("chat_history_conf").
			Default(0).
			Comment("聊天记录配置（0不记录 1仅记录文本 2记录文本和语音）"),
		field.Text("system_prompt").
			Optional().
			Comment("角色设定参数"),
		field.Text("summary_memory").
			Optional().
			Comment("总结记忆"),
		field.String("lang_code").
			MaxLen(10).
			Optional().
			Comment("语言编码"),
		field.String("language").
			MaxLen(10).
			Optional().
			Comment("交互语种"),
		field.Int32("sort").
			Default(0).
			Comment("排序权重"),
		field.Int64("creator").
			Optional().
			Comment("创建者ID"),
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
			Comment("更新者ID"),
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

// Edges of the AgentTemplate.
func (AgentTemplate) Edges() []ent.Edge {
	return nil
}

func (AgentTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_agent_template"},
	}
}
