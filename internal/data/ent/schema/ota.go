package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Ota holds the schema definition for the Ota entity.
type Ota struct {
	ent.Schema
}

// Fields of the Ota.
func (Ota) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("ID"),
		field.String("firmware_name").
			MaxLen(100).
			Optional().
			Comment("固件名称"),
		field.String("type").
			MaxLen(50).
			Optional().
			Comment("固件类型"),
		field.String("version").
			MaxLen(50).
			Optional().
			Comment("版本号"),
		field.Int64("size").
			Optional().
			Comment("文件大小(字节)"),
		field.String("remark").
			MaxLen(500).
			Optional().
			Comment("备注/说明"),
		field.String("firmware_path").
			MaxLen(255).
			Optional().
			Comment("固件路径"),
		field.Int32("sort").
			Default(0).
			Comment("排序"),
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

// Edges of the Ota.
func (Ota) Edges() []ent.Edge {
	return nil
}

func (Ota) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_ota"},
	}
}
