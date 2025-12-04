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

// SysDictType holds the schema definition for the SysDictType entity.
type SysDictType struct {
	ent.Schema
}

// Fields of the SysDictType.
func (SysDictType) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("id"),
		field.String("dict_type").
			MaxLen(100).
			Comment("字典类型"),
		field.String("dict_name").
			MaxLen(255).
			Comment("字典名称"),
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

// Edges of the SysDictType.
func (SysDictType) Edges() []ent.Edge {
	return nil
}

// Indexes of the SysDictType.
func (SysDictType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dict_type").Unique(),
	}
}

func (SysDictType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_dict_type"},
	}
}
