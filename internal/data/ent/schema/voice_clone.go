package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VoiceClone holds the schema definition for the VoiceClone entity.
type VoiceClone struct {
	ent.Schema
}

// Fields of the VoiceClone.
func (VoiceClone) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("唯一标识"),
		field.String("name").
			MaxLen(64).
			Optional().
			Comment("声音名称"),
		field.String("model_id").
			MaxLen(32).
			Optional().
			Comment("模型id"),
		field.String("voice_id").
			MaxLen(32).
			Optional().
			Comment("声音id"),
		field.Int64("user_id").
			Optional().
			Comment("用户ID（关联用户表）"),
		field.Bytes("voice").
			Optional().
			Comment("声音"),
		field.Int32("train_status").
			Default(0).
			Comment("训练状态：0待训练 1训练中 2训练成功 3训练失败"),
		field.String("train_error").
			MaxLen(255).
			Optional().
			Comment("训练错误原因"),
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
	}
}

// Edges of the VoiceClone.
func (VoiceClone) Edges() []ent.Edge {
	return nil
}

// Indexes of the VoiceClone.
func (VoiceClone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("model_id", "user_id", "train_status").
			StorageKey("idx_ai_voice_clone_user_id_model_id_train_status"),
		index.Fields("voice_id").
			StorageKey("idx_ai_voice_clone_voice_id"),
		index.Fields("user_id").
			StorageKey("idx_ai_voice_clone_user_id"),
		index.Fields("model_id", "voice_id").
			StorageKey("idx_ai_voice_clone_model_id_voice_id"),
	}
}

func (VoiceClone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_voice_clone"},
	}
}
