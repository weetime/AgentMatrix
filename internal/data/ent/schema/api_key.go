package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// ApiKey holds the schema definition for the ApiKey entity.
type ApiKey struct {
	ent.Schema
}

// Fields of the ApiKey.
func (ApiKey) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),
		field.UUID("uuid", uuid.UUID{}).
			Unique(),
		field.String("key").
			Default(""),
		field.String("username").
			Default(""),
		field.String("workspace_name").
			Default(""),
		field.String("name").
			Default(""),
		field.String("models").
			Default(""),
		field.Bool("is_enabled").
			Default(true),
		field.Bool("is_deleted").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime",
				dialect.Postgres: "timestamp",
			}),
	}
}

// Edges of the ApiKey.
func (ApiKey) Edges() []ent.Edge {
	return nil
}
