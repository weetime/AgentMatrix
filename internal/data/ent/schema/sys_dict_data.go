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

// SysDictData holds the schema definition for the SysDictData entity.
type SysDictData struct {
	ent.Schema
}

// Fields of the SysDictData.
func (SysDictData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("id"),
		field.Int64("dict_type_id").
			Comment("字典类型ID"),
		field.String("dict_label").
			MaxLen(255).
			Comment("字典标签"),
		field.String("dict_value").
			MaxLen(255).
			Optional().
			Comment("字典值"),
		field.String("remark").
			MaxLen(255).
			Optional().
			Comment("备注"),
		field.Int32("sort").
			Optional().
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

// Edges of the SysDictData.
func (SysDictData) Edges() []ent.Edge {
	return nil
}

// Indexes of the SysDictData.
func (SysDictData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dict_type_id", "dict_value").
			Unique().
			StorageKey("uk_dict_type_value"),
		index.Fields("sort").
			StorageKey("idx_sort"),
	}
}

func (SysDictData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_dict_data"},
	}
}
