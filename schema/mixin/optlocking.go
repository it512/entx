package mixin

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

const lockVerField = "lock_ver"

type optLockInterface interface {
	WhereP(...func(*sql.Selector))
	ResetLockVer()
	LockVer() (int64, bool)
	AddLockVer(int64)
}

type optimisticLockingKey struct{}

// IgnoreOptimisticLocking ingore Optionistic Locking
func IgnoreOptimisticLocking(parent context.Context) context.Context {
	return context.WithValue(parent, optimisticLockingKey{}, true)
}

// OptimisticLocking for ent. db field is lock_ver int64
// only OpUpdateOne do update table set lock_ver = lock_ver +1 , set ... where lock_ver = n -- n is SetLockVer(n)
// when opUpdate OptimisticLocking was ResetLockVer
type OptimisticLocking struct {
	mixin.Schema
}

func (v OptimisticLocking) Fields() []ent.Field {
	return []ent.Field{
		field.Int64(lockVerField).
			Optional().
			Annotations(
				entsql.Default("0"),
			),
	}
}

func (v OptimisticLocking) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(lockVerField),
	}
}

func (v OptimisticLocking) Hooks() []ent.Hook {
	return []ent.Hook{
		optLockingHook(),
	}
}

func optLockingHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op().Is(ent.OpUpdateOne | ent.OpUpdate) {
				if skip, _ := ctx.Value(optimisticLockingKey{}).(bool); skip {
					return next.Mutate(ctx, m)
				}

				mx, ok := m.(optLockInterface)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}

				if m.Op().Is(ent.OpUpdateOne) {
					if val, exists := mx.LockVer(); exists {
						// update table set lock_ver = lock_ver + 1 where lock_ver = val
						mx.ResetLockVer()
						mx.AddLockVer(1)
						addEq(mx, val)
					} // no SetLockVer() , noting to do!
				} else {
					// Op = OpUpdate
					mx.ResetLockVer()
				}
			}

			return next.Mutate(ctx, m)
		})
	}
}

func addEq(w optLockInterface, v int64) {
	w.WhereP(
		sql.FieldEQ(lockVerField, v),
	)
}
