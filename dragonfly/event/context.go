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

// Cancel cancels the context.
func (ctx *Context) Cancel() {
	ctx.cancel = true
}

// Continue calls the function f if the context is not cancelled. If it is cancelled, Continue will return
// immediately.
func (ctx *Context) Continue(f func()) {
	if !ctx.cancel {
		f()
	}
}

// Stop calls the function f if the context is cancelled. If it is not cancelled, Stop will return
// immediately.
// Stop does the opposite of Continue.
func (ctx *Context) Stop(f func()) {
	if ctx.cancel {
		f()
	}
}
