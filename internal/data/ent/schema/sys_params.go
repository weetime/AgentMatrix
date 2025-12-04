package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
)

// SysParams holds the schema definition for the SysParams entity.
type SysParams struct {
	ent.Schema
}

// Fields of the SysParams.
func (SysParams) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.String("param_code").
			MaxLen(100).
			Unique().
			Comment("参数编码"),
		field.String("param_value").
			MaxLen(2000).
			Default("").
			Comment("参数值"),
		field.String("value_type").
			MaxLen(20).
			Default("string").
			Comment("值类型：string-字符串，number-数字，boolean-布尔，array-数组，json-JSON对象"),
		field.Int8("param_type").
			Default(1).
			Comment("类型 0：系统参数 1：非系统参数"),
		field.String("remark").
			MaxLen(200).
			Optional().
			Comment("备注"),
		field.Int64("creator").
			Optional().
			Comment("创建者"),
		field.Time("create_date").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
		field.Int64("updater").
			Optional().
			Comment("更新者"),
		field.Time("update_date").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("更新时间"),
	}
}

// Edges of the SysParams.
func (SysParams) Edges() []ent.Edge {
	return nil
}
