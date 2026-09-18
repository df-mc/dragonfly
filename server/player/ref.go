package player

import (
	"context"
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

// Ref is a stable reference to a player that outlives callbacks. The *Player
// value is only handed to scheduled owner callbacks, where it is safe to use.
type Ref = world.EntityRef[*Player]

// NewRef creates a typed player reference from an entity handle.
func NewRef(h *world.EntityHandle) Ref { return world.NewEntityRef[*Player](h) }

// Do schedules f to run with the player identified by h on its current world owner.
func Do(h *world.EntityHandle, f func(tx *world.Tx, p *Player)) *world.Task {
	return NewRef(h).Do(f)
}

// DoAfter schedules f to run with the player identified by h after delay.
func DoAfter(h *world.EntityHandle, delay time.Duration, f func(tx *world.Tx, p *Player)) *world.Task {
	return NewRef(h).DoAfter(delay, f)
}

// Call runs f with the player identified by h on its current world owner and
// waits for its typed result.
func Call[T any](ctx context.Context, h *world.EntityHandle, f func(tx *world.Tx, p *Player) (T, error)) (T, error) {
	return world.CallRef(ctx, NewRef(h), f)
}

// Do schedules f to run with the player on its current world owner. Use it to
// re-enter the player from code that outlived a callback.
func (p *Player) Do(f func(tx *world.Tx, p *Player)) *world.Task {
	if p == nil {
		return world.NewFinishedTask(world.ErrEntityClosed)
	}
	return Do(p.handle, f)
}

// DoAfter schedules f on the player's current world owner after delay.
func (p *Player) DoAfter(delay time.Duration, f func(tx *world.Tx, p *Player)) *world.Task {
	if p == nil {
		return world.NewFinishedTask(world.ErrEntityClosed)
	}
	return DoAfter(p.handle, delay, f)
}
