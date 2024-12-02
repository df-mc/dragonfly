package event

// Context represents the context of an event. Handlers of an event may call methods on the context to change
// the result of the event.
type Context[T any] struct {
	cancel bool
	val    T
}

// C returns a new event context.
func C[T any](v T) *Context[T] {
	return &Context[T]{val: v}
}

// Val returns the subject of the Context.
func (ctx *Context[T]) Val() T {
	return ctx.val
}

// Cancelled returns whether the context has been cancelled.
func (ctx *Context[T]) Cancelled() bool {
	return ctx.cancel
}

// Cancel cancels the context.
func (ctx *Context[T]) Cancel() {
	ctx.cancel = true
}
