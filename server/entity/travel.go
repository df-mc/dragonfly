package entity

import (
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
)

// PortalTravelComputer handles portal-triggered interdimensional travel for an entity.
type PortalTravelComputer struct {
	// Instantaneous returns true if the entity should skip the portal wait timer. Source is the dimension being left
	// and target is the dimension being entered. Players use this for Creative mode and for End travel (which is
	// always instant in vanilla, regardless of game mode).
	Instantaneous func(source, target world.Dimension) bool
	// Teleport teleports the entity to the final portal position. If nil, Traveller.Teleport is used.
	Teleport func(e Traveller, pos mgl64.Vec3)

	mu             sync.Mutex
	start          time.Time
	inside         bool
	awaitingTravel bool
	travelling     bool
	timedOut       bool
	pending        *world.World
}

// NewPortalTravelComputer creates a PortalTravelComputer for instant portal travel.
func NewPortalTravelComputer() *PortalTravelComputer {
	return &PortalTravelComputer{Instantaneous: func(world.Dimension, world.Dimension) bool { return true }}
}

type portalTravelComputerProvider interface {
	PortalTravelComputer() *PortalTravelComputer
}

// Traveller represents a world.Entity that can travel between dimensions.
type Traveller interface {
	world.Entity
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
}

type portalTravelHandler interface {
	HandlePortalTravel(source, destination world.Dimension)
}

// EnterPortal handles an entity touching a portal block. It teleports the entity to the other dimension after four
// seconds or instantly if instantaneous is true.
func (t *PortalTravelComputer) EnterPortal(e Traveller, tx *world.Tx, target world.Dimension) {
	if destination := t.enterPortal(tx, target); destination != nil {
		t.travelQueued(e, tx, destination)
	}
}

// queuePortalTravel records portal travel to be completed by a terminal Ent tick step.
func (t *PortalTravelComputer) queuePortalTravel(tx *world.Tx, target world.Dimension) {
	if destination := t.enterPortal(tx, target); destination != nil {
		t.mu.Lock()
		t.pending = destination
		t.mu.Unlock()
	}
}

// enterPortal updates portal contact state and returns the destination world if travel should start.
func (t *PortalTravelComputer) enterPortal(tx *world.Tx, target world.Dimension) *world.World {
	source := tx.World()
	destination := source.PortalDestination(target)
	if destination == source {
		return nil
	}

	t.mu.Lock()
	t.inside = true
	if t.timedOut {
		// Timed out, we can't travel through portals.
		t.mu.Unlock()
		return nil
	}
	travelNow := t.instantaneous(source.Dimension(), target) || (t.awaitingTravel && time.Since(t.start) >= time.Second*4)
	if !travelNow && !t.awaitingTravel {
		t.start, t.awaitingTravel = time.Now(), true
	}
	t.mu.Unlock()

	if travelNow {
		return destination
	}
	return nil
}

func (t *PortalTravelComputer) instantaneous(source, target world.Dimension) bool {
	return t.Instantaneous != nil && t.Instantaneous(source, target)
}

// hasPendingPortalTravel reports whether portal travel was queued during this tick.
func (t *PortalTravelComputer) hasPendingPortalTravel() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pending != nil
}

// finishPendingPortalTravel consumes queued portal travel and starts the terminal transfer.
func (t *PortalTravelComputer) finishPendingPortalTravel(e Traveller, tx *world.Tx) bool {
	t.mu.Lock()
	destination := t.pending
	t.pending = nil
	t.mu.Unlock()

	if destination == nil {
		return false
	}
	t.travel(e, tx, destination)
	return true
}

// StopPortalContact resets the portal timer if the entity was not inside a portal this tick.
func (t *PortalTravelComputer) StopPortalContact() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inside {
		t.inside = false
		return
	}
	if t.travelling || t.pending != nil {
		return
	}
	t.timedOut, t.awaitingTravel = false, false
}

// travel removes the entity from the current world and queues it for the given Nether or Overworld world.
func (t *PortalTravelComputer) travel(e Traveller, tx *world.Tx, destination *world.World) {
	source := tx.World()
	if destination == nil || destination == source {
		return
	}

	sourceDim, destinationDim := source.Dimension(), destination.Dimension()
	pos := translatePortalPosition(cube.PosFromVec3(e.Position()), sourceDim, destinationDim)

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	handle := tx.RemoveEntity(e)
	if handle == nil {
		t.mu.Lock()
		t.travelling = false
		t.mu.Unlock()
		return
	}

	go func() {
		<-destination.Exec(func(tx *world.Tx) {
			spawn := arrivalSpawn(tx, sourceDim, destinationDim, pos)

			if e, ok := tx.AddEntityAt(handle, spawn).(Traveller); ok {
				t.finishTravel(e, spawn, sourceDim, destinationDim)
			}
		})

		t.mu.Lock()
		t.travelling = false
		t.mu.Unlock()
	}()
}

// travelQueued moves the entity after the current transaction finishes. This is used by callers such as players that
// may touch a portal from the middle of a tick and continue running afterwards.
func (t *PortalTravelComputer) travelQueued(e Traveller, tx *world.Tx, destination *world.World) {
	source := tx.World()
	if destination == nil || destination == source {
		return
	}

	sourceDim, destinationDim := source.Dimension(), destination.Dimension()
	pos := translatePortalPosition(cube.PosFromVec3(e.Position()), sourceDim, destinationDim)

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	go func() {
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
			spawn := arrivalSpawn(tx, sourceDim, destinationDim, pos)

			if e, ok := tx.AddEntityAt(handle, spawn).(Traveller); ok {
				t.finishTravel(e, spawn, sourceDim, destinationDim)
			}
		})

		t.mu.Lock()
		t.travelling = false
		t.mu.Unlock()
	}()
}

// arrivalSpawn computes the entity arrival position in the destination world and performs any required side effects
// (nether portal find-or-create, End spawn-platform regeneration). The fallback position is used when no
// dimension-specific structure exists.
func arrivalSpawn(tx *world.Tx, sourceDim, destinationDim world.Dimension, fallback cube.Pos) mgl64.Vec3 {
	switch destinationDim {
	case world.End:
		portal.GenerateEndSpawnPlatform(tx)
		return portal.EndSpawnPosition()
	case world.Nether:
		if n, ok := portal.FindOrCreateNetherPortal(tx, fallback, 128); ok {
			return n.Spawn().Vec3Middle()
		}
	case world.Overworld:
		if sourceDim == world.End {
			return tx.World().Spawn().Vec3Middle()
		}
		if n, ok := portal.FindOrCreateNetherPortal(tx, fallback, 128); ok {
			return n.Spawn().Vec3Middle()
		}
	}
	return fallback.Vec3Middle()
}

// finishTravel runs the post-transfer portal hook and places the traveller at
// the destination spawn position.
func (t *PortalTravelComputer) finishTravel(e Traveller, pos mgl64.Vec3, source, destination world.Dimension) {
	handlePortalTravel(e, source, destination)
	if t.Teleport != nil {
		t.Teleport(e, pos)
		return
	}
	e.Teleport(pos)
}

// handlePortalTravel dispatches portal travel hooks to Ent behaviours and
// non-Ent travellers that implement portalTravelHandler.
func handlePortalTravel(e Traveller, source, destination world.Dimension) {
	if ent, ok := e.(*Ent); ok {
		if h, ok := ent.Behaviour().(portalTravelHandler); ok {
			h.HandlePortalTravel(source, destination)
		}
		return
	}
	if h, ok := e.(portalTravelHandler); ok {
		h.HandlePortalTravel(source, destination)
	}
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
