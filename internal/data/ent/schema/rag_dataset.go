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

// RagDataset holds the schema definition for the RagDataset entity.
type RagDataset struct {
	ent.Schema
}

// Fields of the RagDataset.
func (RagDataset) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("唯一标识"),
		field.String("dataset_id").
			MaxLen(64).
			Comment("知识库ID"),
		field.String("rag_model_id").
			MaxLen(64).
			Optional().
			Comment("RAG模型配置ID"),
		field.String("name").
			MaxLen(100).
			Comment("知识库名称"),
		field.Text("description").
			Optional().
			Comment("知识库描述"),
		field.Int32("status").
			Default(1).
			Comment("状态：0停用 1启用"),
		field.Int64("creator").
			Optional().
			Comment("创建者ID"),
		field.Time("created_at").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
		field.Int64("updater").
			Optional().
			Comment("更新者ID"),
		field.Time("updated_at").
			Optional().
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("更新时间"),
	}
}

// Edges of the RagDataset.
func (RagDataset) Edges() []ent.Edge {
	return nil
}

// Indexes of the RagDataset.
func (RagDataset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dataset_id").
			Unique().
			StorageKey("uk_dataset_id"),
		index.Fields("status").
			StorageKey("idx_ai_rag_dataset_status"),
		index.Fields("creator").
			StorageKey("idx_ai_rag_dataset_creator"),
		index.Fields("created_at").
			StorageKey("idx_ai_rag_dataset_created_at"),
	}
}

func (RagDataset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_rag_dataset"},
	}
}
