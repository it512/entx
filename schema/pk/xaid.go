package schematype

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/it512/xaid"
)

type XAIDPk struct {
	mixin.Schema
}

func (XAIDPk) Fields() []ent.Field {
	return []ent.Field{
		XAIDField("id"),
	}
}

func XAIDField(colname string) ent.Field {
	return field.String(colname).
		MaxLen(xaid.Length).
		MinLen(xaid.Length).
		SchemaType(map[string]string{
			dialect.Postgres: "char(27)",
			dialect.MySQL:    "char(27)",
		}).
		GoType(xaid.Nil).
		DefaultFunc(xaid.NewOrNil)
}
