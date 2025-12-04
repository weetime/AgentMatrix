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

// ModelProvider holds the schema definition for the ModelProvider entity.
type ModelProvider struct {
	ent.Schema
}

// Fields of the ModelProvider.
func (ModelProvider) Fields() []ent.Field {
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
		field.String("provider_code").
			MaxLen(50).
			Optional().
			Comment("供应器类型"),
		field.String("name").
			MaxLen(50).
			Optional().
			Comment("供应器名称"),
		field.String("fields").
			SchemaType(map[string]string{
				dialect.MySQL:    "json",
				dialect.Postgres: "jsonb",
			}).
			Optional().
			Comment("供应器字段列表(JSON格式)"),
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

// Edges of the ModelProvider.
func (ModelProvider) Edges() []ent.Edge {
	return nil
}

// Indexes of the ModelProvider.
func (ModelProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("model_type").
			StorageKey("idx_ai_model_provider_model_type"),
	}
}

func (ModelProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_model_provider"},
	}
}
