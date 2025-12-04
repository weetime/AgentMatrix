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

// SysUserToken holds the schema definition for the SysUserToken entity.
type SysUserToken struct {
	ent.Schema
}

// Fields of the SysUserToken.
func (SysUserToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("id"),
		field.Int64("user_id").
			Comment("用户id"),
		field.String("token").
			MaxLen(100).
			Comment("用户token"),
		field.Time("expire_date").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("过期时间"),
		field.Time("update_date").
			Optional().
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("更新时间"),
		field.Time("create_date").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("创建时间"),
	}
}

// Edges of the SysUserToken.
func (SysUserToken) Edges() []ent.Edge {
	return nil
}

// Indexes of the SysUserToken.
func (SysUserToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").
			Unique().
			StorageKey("user_id"),
		index.Fields("token").
			Unique().
			StorageKey("token"),
	}
}

func (SysUserToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user_token"},
	}
}
