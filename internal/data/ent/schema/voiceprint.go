package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Voiceprint holds the schema definition for the Voiceprint entity.
type Voiceprint struct {
	ent.Schema
}

// Fields of the Voiceprint.
func (Voiceprint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("声纹唯一标识"),
		field.String("name").
			MaxLen(64).
			Optional().
			Comment("声纹名称"),
		field.Int64("user_id").
			Optional().
			Comment("用户ID（关联用户表）"),
		field.String("agent_id").
			MaxLen(32).
			Optional().
			Comment("关联智能体ID"),
		field.String("agent_code").
			MaxLen(36).
			Optional().
			Comment("关联智能体编码"),
		field.String("agent_name").
			MaxLen(36).
			Optional().
			Comment("关联智能体名称"),
		field.String("description").
			MaxLen(255).
			Optional().
			Comment("声纹描述"),
		field.Text("embedding").
			Optional().
			Comment("声纹特征向量（JSON数组格式）"),
		field.Text("memory").
			Optional().
			Comment("关联记忆数据"),
		field.Int32("sort").
			Default(0).
			Comment("排序权重"),
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

// Edges of the Voiceprint.
func (Voiceprint) Edges() []ent.Edge {
	return nil
}

func (Voiceprint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_voiceprint"},
	}
}
