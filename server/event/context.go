package event

// Context represents the context of an event. Handlers of an event may call methods on the context to change
// the result of the event.
type Context struct {
	cancel bool
}

// C returns a new event context.
func C() *Context {
	return &Context{}
}

// Cancelled returns whether the context has been cancelled.
func (ctx *Context) Cancelled() bool {
	return ctx.cancel
}

// Cancel cancels the context.
func (ctx *Context) Cancel() {
	ctx.cancel = true
}
