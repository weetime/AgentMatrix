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

// Device holds the schema definition for the Device entity.
type Device struct {
	ent.Schema
}

// Fields of the Device.
func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			Unique().
			Immutable().
			Comment("设备唯一标识"),
		field.Int64("user_id").
			Optional().
			Comment("关联用户ID"),
		field.String("mac_address").
			MaxLen(50).
			Optional().
			Comment("MAC地址"),
		field.Time("last_connected_at").
			Optional().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}).
			Comment("最后连接时间"),
		field.Int32("auto_update").
			Default(0).
			Comment("自动更新开关(0关闭/1开启)"),
		field.String("board").
			MaxLen(50).
			Optional().
			Comment("设备硬件型号"),
		field.String("alias").
			MaxLen(64).
			Optional().
			Comment("设备别名"),
		field.String("agent_id").
			MaxLen(32).
			Optional().
			Comment("智能体ID"),
		field.String("app_version").
			MaxLen(20).
			Optional().
			Comment("固件版本号"),
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

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return nil
}

// Indexes of the Device.
func (Device) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mac_address").
			StorageKey("idx_ai_device_created_at"),
	}
}

func (Device) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ai_device"},
	}
}
