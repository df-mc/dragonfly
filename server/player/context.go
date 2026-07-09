package player

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Context is the context passed to player event callbacks. It embeds the
// world Context, so world operations and Cancel are available directly, and
// adds the Player the event concerns. It is valid only during the callback.
type Context struct {
	*world.Context
	p *Player
}

// newContext returns a Context for one event dispatch concerning p.
func newContext(p *Player) *Context {
	return &Context{Context: p.tx.Event(), p: p}
}

// Player returns the player the event concerns, valid only during the
// callback.
func (ctx *Context) Player() *Player { return ctx.p }

// Defer schedules f to run on the owner after the current callback completes,
// with the player re-resolved for that moment. If the player left the world by
// then, the task fails with world.ErrEntityClosed.
func (ctx *Context) Defer(f func(ctx *Context)) *world.Task {
	h := ctx.p.H()
	return ctx.DeferErr(func(wctx *world.Context) error {
		if e, ok := h.Entity(wctx); ok {
			f(newContext(e.(*Player)))
			return nil
		}
		return world.ErrEntityClosed
	})
}
