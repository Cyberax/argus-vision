package utils

import (
	"context"
	"sync"
)

var mutableContextKey = "mutableContext"

type MutableContextKey[V any] struct {
	name string
}

type MutableContext struct {
	previousContext *MutableContext

	mtx    sync.Mutex
	values map[any]any
}

func NewMutableContextKey[V any](name string) MutableContextKey[V] {
	return MutableContextKey[V]{
		name: name,
	}
}

func mustGetMutableContext(ctx context.Context) *MutableContext {
	res := tryGetMutableContext(ctx)
	PanicIfF(res == nil, "no mutable context found")
	return res
}

func tryGetMutableContext(ctx context.Context) *MutableContext {
	value, ok := ctx.Value(mutableContextKey).(*MutableContext)
	if !ok {
		return nil
	}
	return value
}

func ForkMutableContext(ctx context.Context) context.Context {
	newMutCtx := &MutableContext{
		values: map[any]any{},
	}

	prevCtx, ok := ctx.Value(mutableContextKey).(*MutableContext)
	if ok {
		newMutCtx.previousContext = prevCtx
	}

	return context.WithValue(ctx, mutableContextKey, newMutCtx)
}

func SetMutableContextValue[V any](ctx context.Context, key MutableContextKey[V], value V) {
	m := mustGetMutableContext(ctx)

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.values[key] = value
}

type EditorFunc[V any] func(value V, present bool) (newValue V, newPresent bool)

func EditMutableContextValue[V any](ctx context.Context, key MutableContextKey[V], editor EditorFunc[V]) {
	m := mustGetMutableContext(ctx)

	m.mtx.Lock()
	defer m.mtx.Unlock()

	cur, present := m.values[key].(V)
	newVal, newPresent := editor(cur, present)

	if newPresent {
		m.values[key] = newVal
	} else {
		delete(m.values, key)
	}
}

func TryGetMutableContextValue[V any](ctx context.Context, key MutableContextKey[V]) (V, bool) {
	var none V

	tryGet := func(curM *MutableContext) (V, bool) {
		curM.mtx.Lock()
		defer curM.mtx.Unlock()
		res, ok := curM.values[key]
		if ok {
			return res.(V), true
		}
		return none, false
	}

	m := mustGetMutableContext(ctx)

	for m != nil {
		res, ok := tryGet(m)
		if ok {
			return res, true
		}
		// The previousContext member is never mutated, so it's safe to read it outside
		// the locked area.
		m = m.previousContext
	}

	return none, false
}

func MustGetMutableContextValue[V any](ctx context.Context, key MutableContextKey[V]) V {
	res, ok := TryGetMutableContextValue(ctx, key)
	PanicIfF(!ok, "failed to find the value")
	return res
}
