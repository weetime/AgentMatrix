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

// ModelConfig holds the schema definition for the ModelConfig entity.
type ModelConfig struct {
	ent.Schema
}

// Fields of the ModelConfig.
func (ModelConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("主键"),
		field.String("model_type").
			MaxLen(20).
			Optional().
			Comment("模型类型(Memory/ASR/VAD/LLM/TTS)"),
		field.String("model_code").
			MaxLen(50).
			Optional().
			Comment("模型编码(如AliLLM、DoubaoTTS)"),
		field.String("model_name").
			MaxLen(50).
			Optional().
			Comment("模型名称"),
		field.Bool("is_default").
			Default(false).
			Comment("是否默认配置(0否 1是)"),
		field.Bool("is_enabled").
			Default(false).
			Comment("是否启用"),
		field.String("config_json").
			SchemaType(map[string]string{
				dialect.MySQL:    "json",
				dialect.Postgres: "jsonb",
			}).
			Optional().
			Comment("模型配置(JSON格式)"),
		field.String("doc_link").
			MaxLen(200).
			Optional().
			Comment("官方文档链接"),
		field.String("remark").
			MaxLen(255).
			Optional().
			Comment("备注"),
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

// Edges of the ModelConfig.
func (ModelConfig) Edges() []ent.Edge {
	return nil
}

// Indexes of the ModelConfig.
func (ModelConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("model_type").
			StorageKey("idx_ai_model_config_model_type"),
	}
}

func (ModelConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_model_config"},
	}
}
