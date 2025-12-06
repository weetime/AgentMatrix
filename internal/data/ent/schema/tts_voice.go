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

// TtsVoice holds the schema definition for the TtsVoice entity.
type TtsVoice struct {
	ent.Schema
}

// Fields of the TtsVoice.
func (TtsVoice) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("主键"),
		field.String("tts_model_id").
			MaxLen(32).
			Optional().
			Comment("对应TTS模型主键"),
		field.String("name").
			MaxLen(20).
			Optional().
			Comment("音色名称"),
		field.String("tts_voice").
			MaxLen(50).
			Optional().
			Comment("音色编码"),
		field.String("languages").
			MaxLen(50).
			Optional().
			Comment("语言"),
		field.String("voice_demo").
			MaxLen(500).
			Optional().
			Comment("音色Demo"),
		field.String("remark").
			MaxLen(255).
			Optional().
			Comment("备注"),
		field.String("reference_audio").
			MaxLen(500).
			Optional().
			Comment("参考音频路径"),
		field.String("reference_text").
			MaxLen(500).
			Optional().
			Comment("参考文本"),
		field.Int32("sort").
			Default(0).
			Comment("排序"),
		field.Int64("creator").
			Optional().
			Comment("创建者ID"),
		field.Time("create_date").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
		field.Int64("updater").
			Optional().
			Comment("更新者ID"),
		field.Time("update_date").
			Optional().
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("更新时间"),
	}
}

// Edges of the TtsVoice.
func (TtsVoice) Edges() []ent.Edge {
	return nil
}

// Indexes of the TtsVoice.
func (TtsVoice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tts_model_id").
			StorageKey("idx_ai_tts_voice_tts_model_id"),
	}
}

func (TtsVoice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_tts_voice"},
	}
}
