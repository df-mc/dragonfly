package entity

import (
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
)

// TravelComputer handles the interdimensional travelling of an entity.
type TravelComputer struct {
	// Instantaneous is a function that returns true if the entity given can travel instantly.
	Instantaneous func() bool

	mu             sync.RWMutex
	start          time.Time
	inside         bool
	awaitingTravel bool
	travelling     bool
	timedOut       bool
}

// Traveller represents a world.Entity that can travel between dimensions.
type Traveller interface {
	world.Entity
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
}

// EnterPortal handles an entity touching a portal block. It teleports the entity to the other dimension after four
// seconds or instantly if instantaneous is true.
func (t *TravelComputer) EnterPortal(travel Traveller, tx *world.Tx, target world.Dimension) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.inside = true
	switch target {
	case world.Nether:
		if t.timedOut {
			// Timed out, we can't travel through Nether portals.
			return
		}
		if t.Instantaneous() || (t.awaitingTravel && time.Since(t.start) >= time.Second*4) {
			t.mu.Unlock()
			t.Travel(travel, tx.World(), tx.World().PortalDestination(world.Nether))
			t.mu.Lock()
		} else if !t.awaitingTravel {
			t.start, t.awaitingTravel = time.Now(), true
		}
	}
}

// StopTravelling resets the travel timer if the entity was not inside a portal this tick.
func (t *TravelComputer) StopTravelling() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inside {
		t.inside = false
		return
	}
	t.inside = false
	if t.travelling {
		return
	}
	t.timedOut, t.awaitingTravel = false, false
}

// Travel moves the player to the given Nether or Overworld world and translates the player's current position based
// on the source world.
func (t *TravelComputer) Travel(e Traveller, source *world.World, destination *world.World) {
	sourceDimension, targetDimension := source.Dimension(), destination.Dimension()
	pos := cube.PosFromVec3(e.Position())
	switch sourceDimension {
	case world.Overworld:
		pos = cube.Pos{pos.X() / 8, pos.Y() + sourceDimension.Range().Min(), pos.Z() / 8}
	case world.Nether:
		pos = cube.Pos{pos.X() * 8, pos.Y() - targetDimension.Range().Min(), pos.Z() * 8}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false

	go func() {
		spawn := pos.Vec3Middle()

		source.Exec(func(tx *world.Tx) {
			tx.RemoveEntity(e)
		})

		destination.Exec(func(tx *world.Tx) {
			if netherPortal, ok := portal.FindOrCreateNetherPortal(tx, pos, 128); ok {
				spawn = netherPortal.Spawn().Vec3Middle()
			}

			tx.AddEntity(e.H())
			if ent, ok := e.H().Entity(tx); ok {
				ent.(Traveller).Teleport(spawn)
			}
		})

		t.mu.Lock()
		defer t.mu.Unlock()
		t.travelling = false
	}()
}
