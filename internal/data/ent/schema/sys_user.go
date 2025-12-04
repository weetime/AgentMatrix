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

// SysUser holds the schema definition for the SysUser entity.
type SysUser struct {
	ent.Schema
}

// Fields of the SysUser.
func (SysUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("id"),
		field.String("username").
			MaxLen(50).
			Comment("用户名"),
		field.String("password").
			MaxLen(100).
			Optional().
			Comment("密码"),
		field.Int32("super_admin").
			Optional().
			Comment("超级管理员   0：否   1：是"),
		field.Int32("status").
			Optional().
			Comment("状态  0：停用   1：正常"),
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
		field.Int64("creator").
			Optional().
			Comment("创建者ID"),
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

// Edges of the SysUser.
func (SysUser) Edges() []ent.Edge {
	return nil
}

// Indexes of the SysUser.
func (SysUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").
			Unique().
			StorageKey("uk_username"),
	}
}

func (SysUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user"},
	}
}
