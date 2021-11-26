package event

// Context represents the context of an event. Handlers of an event may call methods on the context to change
// the result of the event.
type Context struct {
	cancel bool
	after  []func(bool)
}

// C returns a new event context.
func C() *Context {
	return &Context{}
}

// Cancel cancels the context.
func (ctx *Context) Cancel() {
	ctx.cancel = true
}

// After calls the function passed after the action of the event has been completed, either by a call to
// (*Context).Continue() or (*Context).Stop().
// After can be executed multiple times to attach more functions to be called after the event is executed.
func (ctx *Context) After(f func(cancelled bool)) {
	ctx.after = append(ctx.after, f)
}

// Continue calls the function f if the context is not cancelled. If it is cancelled, Continue will return
// immediately.
// These functions are not generally useful for handling events. See After() for executing code after the
// event happens.
func (ctx *Context) Continue(f func()) {
	if !ctx.cancel {
		f()
		for _, v := range ctx.after {
			v(ctx.cancel)
		}
	}
}

// Stop calls the function f if the context is cancelled. If it is not cancelled, Stop will return
// immediately.
// Stop does the opposite of Continue.
// These functions are not generally useful for handling events. See After() for executing code after the
// event happens.
func (ctx *Context) Stop(f func()) {
	if ctx.cancel {
		f()
		for _, v := range ctx.after {
			v(ctx.cancel)
		}
	}
}
