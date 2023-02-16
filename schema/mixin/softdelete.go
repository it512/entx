package mixin

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type SoftDelete struct {
	mixin.Schema
}

func (SoftDelete) Fields() []ent.Field {
	return []ent.Field{
		field.Time("delete_time").
			Optional(),
	}
}

func (SoftDelete) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("delete_time"),
	}
}

type softDeleteKey struct{}

func IgnoreSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

type traverseFunc func(context.Context, x) error

func (f traverseFunc) Intercept(next ent.Querier) ent.Querier {
	return next
}

func (d SoftDelete) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		traverseFunc(func(ctx context.Context, q x) error {
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			d.p(q)
			return nil
		}),
	}
}

func (d SoftDelete) Hooks() []ent.Hook {
	return []ent.Hook{
		on(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					mx, ok := m.(a)
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T", m)
					}
					d.p(mx)
					mx.SetOp(ent.OpUpdate)
					mx.SetDeleteTime(time.Now())
					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

// p adds a storage-level predicate to the queries and mutations.
func (d SoftDelete) p(w x) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}

type x interface {
	WhereP(...func(*sql.Selector))
}

type a interface {
	x
	SetOp(ent.Op)
	Client() ent.Mutator
	SetDeleteTime(time.Time)
}

func on(hk ent.Hook, op ent.Op) ent.Hook {
	return iif(hk, hasOp(op))
}

func hasOp(op ent.Op) condition {
	return func(_ context.Context, m ent.Mutation) bool {
		return m.Op().Is(op)
	}
}

func iif(hk ent.Hook, cond condition) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if cond(ctx, m) {
				return hk(next).Mutate(ctx, m)
			}
			return next.Mutate(ctx, m)
		})
	}
}

type condition func(context.Context, ent.Mutation) bool
