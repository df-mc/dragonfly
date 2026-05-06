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
	// Instantaneous returns true if the entity should skip the portal wait timer. Players use this for Creative mode.
	Instantaneous func() bool

	mu             sync.Mutex
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
func (t *TravelComputer) EnterPortal(e Traveller, tx *world.Tx, target world.Dimension) {
	source := tx.World()
	destination := source.PortalDestination(target)
	if destination == source {
		return
	}

	t.mu.Lock()
	t.inside = true
	if t.timedOut {
		// Timed out, we can't travel through portals.
		t.mu.Unlock()
		return
	}
	travelNow := t.instantaneous() || (t.awaitingTravel && time.Since(t.start) >= time.Second*4)
	if !travelNow && !t.awaitingTravel {
		t.start, t.awaitingTravel = time.Now(), true
	}
	t.mu.Unlock()

	if travelNow {
		t.travel(e, source, destination)
	}
}

func (t *TravelComputer) instantaneous() bool {
	return t.Instantaneous != nil && t.Instantaneous()
}

// StopTravelling resets the travel timer if the entity was not inside a portal this tick.
func (t *TravelComputer) StopTravelling() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inside {
		t.inside = false
		return
	}
	if t.travelling {
		return
	}
	t.timedOut, t.awaitingTravel = false, false
}

// travel moves the entity to the given Nether or Overworld world and translates its current position based on the
// source world.
func (t *TravelComputer) travel(e Traveller, source *world.World, destination *world.World) {
	if destination == nil || destination == source {
		return
	}

	pos := translatePortalPosition(cube.PosFromVec3(e.Position()), source.Dimension(), destination.Dimension())

	t.mu.Lock()
	defer t.mu.Unlock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false

	go func() {
		spawn := pos.Vec3Middle()

		var handle *world.EntityHandle
		<-source.Exec(func(tx *world.Tx) {
			handle = tx.RemoveEntity(e)
		})
		if handle == nil {
			t.mu.Lock()
			t.travelling = false
			t.mu.Unlock()
			return
		}

		<-destination.Exec(func(tx *world.Tx) {
			if netherPortal, ok := portal.FindOrCreateNetherPortal(tx, pos, 128); ok {
				spawn = netherPortal.Spawn().Vec3Middle()
			}

			if ent, ok := tx.AddEntity(handle).(Traveller); ok {
				ent.Teleport(spawn)
			}
		})

		t.mu.Lock()
		defer t.mu.Unlock()
		t.travelling = false
	}()
}

// translatePortalPosition maps a position in the source dimension to the equivalent position in the target dimension.
// Overworld coordinates are divided by 8 when crossing to the Nether and Nether coordinates are multiplied by 8 when
// crossing to the Overworld; the Y coordinate is clamped to the target dimension's vertical range.
func translatePortalPosition(pos cube.Pos, source, target world.Dimension) cube.Pos {
	switch source {
	case world.Overworld:
		pos[0], pos[2] = pos[0]>>3, pos[2]>>3
	case world.Nether:
		pos[0], pos[2] = pos[0]*8, pos[2]*8
	}
	r := target.Range()
	pos[1] = min(max(pos[1], r.Min()), r.Max())
	return pos
}
