package player

import (
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

// Ref is a stable reference to a player that outlives callbacks. The *Player
// value is only handed to scheduled owner callbacks, where it is safe to use.
type Ref = world.EntityRef[*Player]

// NewRef creates a typed player reference from an entity handle.
func NewRef(h *world.EntityHandle) Ref { return world.NewEntityRef[*Player](h) }

// Do schedules f to run with the player on its current world owner. Use it to
// re-enter the player from code that outlived a callback.
func (p *Player) Do(f func(tx *world.Tx, p *Player)) *world.Task {
	if p == nil {
		return world.NewFinishedTask(world.ErrEntityClosed)
	}
	return NewRef(p.handle).Do(f)
}

// DoAfter schedules f on the player's current world owner after delay.
func (p *Player) DoAfter(delay time.Duration, f func(tx *world.Tx, p *Player)) *world.Task {
	if p == nil {
		return world.NewFinishedTask(world.ErrEntityClosed)
	}
	return NewRef(p.handle).DoAfter(delay, f)
}
