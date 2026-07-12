package world

import (
	"context"
	"fmt"
	"time"
)

// EntityRef is a stable, typed reference to an entity. The entity value T is
// only handed to scheduled owner callbacks, where it is safe to use.
type EntityRef[T Entity] struct {
	h *EntityHandle
}

// NewEntityRef creates a typed reference from an EntityHandle.
func NewEntityRef[T Entity](h *EntityHandle) EntityRef[T] { return EntityRef[T]{h: h} }

// Handle returns the underlying stable entity handle.
func (r EntityRef[T]) Handle() *EntityHandle { return r.h }

// Do schedules f on the entity's current world owner, like EntityHandle.Do,
// but hands f the entity as T. If the entity is no longer a T when the task
// runs, the task fails with ErrEntityType.
func (r EntityRef[T]) Do(f func(tx *Tx, e T)) *Task {
	return r.h.schedule(typed(f))
}

// DoAfter schedules f on the entity's world owner after delay, typed like Do.
func (r EntityRef[T]) DoAfter(delay time.Duration, f func(tx *Tx, e T)) *Task {
	return r.h.scheduleAfter(delay, typed(f))
}

// typed wraps f so the scheduled entity is asserted to T before f runs.
func typed[T Entity](f func(tx *Tx, e T)) func(*Tx, Entity) error {
	return func(tx *Tx, e Entity) error {
		v, err := assertEntity[T](e)
		if err != nil {
			return err
		}
		f(tx, v)
		return nil
	}
}

// CallRef runs f with the ref's entity on its current world owner and waits
// for the typed result. Off-owner code only, like Call. If f panics, CallRef
// re-panics with the original value on the waiting goroutine. Context
// cancellation stops pending work, but CallRef waits for a callback that has
// already started.
func CallRef[T any, E Entity](ctx context.Context, ref EntityRef[E], f func(tx *Tx, e E) (T, error)) (T, error) {
	var zero T
	ctx, err := callContext(ctx)
	if err != nil {
		return zero, err
	}
	var result T
	task := ref.h.schedule(func(tx *Tx, e Entity) error {
		v, err := assertEntity[E](e)
		if err != nil {
			return err
		}
		var callErr error
		result, callErr = f(tx, v)
		return callErr
	})
	return awaitTask(ctx, task, &result)
}

// assertEntity converts e to T, returning ErrEntityType if it no longer is one.
func assertEntity[T Entity](e Entity) (T, error) {
	v, ok := e.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("%w: got %T", ErrEntityType, e)
	}
	return v, nil
}
