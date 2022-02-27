package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
	"time"
)

// TravelComputer handles the interdimensional travelling of an entity.
type TravelComputer struct {
	// Instantaneous is a function that returns true if the entity given can travel instantly.
	Instantaneous func() bool

	mu             sync.RWMutex
	start          time.Time
	awaitingTravel bool
	travelling     bool
	timedOut       bool
}

// Traveller represents a world.Entity that can travel between dimensions.
type Traveller interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)

	world.Entity
}

// TickTravelling checks if the player is colliding with a nether portal block. If so, it teleports the player
// to the other dimension after four seconds or instantly if instantaneous is true.
func (t *TravelComputer) TickTravelling(w *world.World, e Traveller) {
	aabb := e.AABB().Translate(e.Position())
	if w.Dimension() == world.Overworld || w.Dimension() == world.Nether {
		// Get all blocks that could touch the player and check if any of them intersect with a portal block.
		for _, pos := range w.BlocksAround(aabb) {
			b := w.Block(pos)
			if p, ok := b.(interface {
				// Portal returns the dimension the portal block takes you to.
				Portal() world.Dimension
			}); ok && p.Portal() == world.Nether {
				for _, a := range b.Model().AABB(pos, w) {
					if a.Translate(pos.Vec3()).IntersectsWith(aabb.Grow(0.25)) {
						t.mu.Lock()
						timeOut, awaitingTravel, start := t.timedOut, t.awaitingTravel, t.start
						t.mu.Unlock()

						if !timeOut {
							if t.Instantaneous() || (awaitingTravel && time.Since(start) >= time.Second*4) {
								d, _ := w.PortalDestinations()
								t.Travel(e, w, d)
							} else if !awaitingTravel {
								t.mu.Lock()
								t.start, t.awaitingTravel = time.Now(), true
								t.mu.Unlock()
							}
						}
						return
					}
				}
			}
		}

		// No portals found. Check if we aren't travelling and if so, reset.
		t.mu.Lock()
		defer t.mu.Unlock()
		if !t.travelling {
			t.timedOut, t.awaitingTravel = false, false
		}
	}
}

// Travel moves the player to the given Nether or Overworld world, and translates the player's current position based
// on the source world.
func (t *TravelComputer) Travel(e Traveller, source *world.World, destination *world.World) {
	sourceDimension, targetDimension := source.Dimension(), destination.Dimension()
	pos := cube.PosFromVec3(e.Position())
	if sourceDimension == world.Overworld {
		pos = cube.Pos{pos.X() / 8, pos.Y() + sourceDimension.Range().Min(), pos.Z() / 8}
	} else if sourceDimension == world.Nether {
		pos = cube.Pos{pos.X() * 8, pos.Y() - targetDimension.Range().Min(), pos.Z() * 8}
	}

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	go func() {
		// Java edition spawns the player at the translated position if all else fails, so we do the same.
		spawn := pos.Vec3Middle()
		if netherPortal, ok := portal.FindOrCreateNetherPortal(destination, pos, 128); ok {
			spawn = netherPortal.Spawn().Vec3Middle()
		}

		// Add the entity to the destination dimension and stop the portal travel status.
		destination.AddEntity(e)
		e.Teleport(spawn)

		t.mu.Lock()
		t.travelling = false
		t.mu.Unlock()
	}()
}
