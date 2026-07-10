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
	// Instantaneous returns true if the entity should skip the portal wait timer. Players use this for game modes
	// with instant portal travel.
	Instantaneous func() bool
	// Teleport teleports the entity to the final portal position. If nil, Traveller.Teleport is used.
	Teleport func(e Traveller, pos mgl64.Vec3)
	// CreatePortal specifies if the entity may create a portal at the destination when none is found. Only players
	// create portals; other entities only travel through portals that are already linked.
	CreatePortal bool
	// Cooldown is how long the entity must wait after a travel attempt before it may travel again. Non-player
	// entities use 15 seconds (300 ticks).
	Cooldown time.Duration

	mu             sync.Mutex
	start          time.Time
	cooldownUntil  time.Time
	inside         bool
	awaitingTravel bool
	travelling     bool
	timedOut       bool
	pending        *world.World
}

// NewPortalTravelComputer creates a PortalTravelComputer for instant portal travel.
func NewPortalTravelComputer() *PortalTravelComputer {
	return &PortalTravelComputer{Instantaneous: func() bool { return true }, Cooldown: time.Second * 15}
}

// portalSearchRadius is the radius around the scaled arrival position searched for an existing linked portal.
// Bedrock Edition searches 128 blocks in both dimensions.
const portalSearchRadius = 128

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
	if time.Now().Before(t.cooldownUntil) {
		t.mu.Unlock()
		return nil
	}
	travelNow := t.instantaneous() || (t.awaitingTravel && time.Since(t.start) >= time.Second*4)
	if !travelNow && !t.awaitingTravel {
		t.start, t.awaitingTravel = time.Now(), true
	}
	t.mu.Unlock()

	if travelNow {
		return destination
	}
	return nil
}

func (t *PortalTravelComputer) instantaneous() bool {
	return t.Instantaneous != nil && t.Instantaneous()
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
	origin := e.Position()
	pos := translatePortalPosition(cube.PosFromVec3(origin), sourceDim, destinationDim)

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	handle := tx.RemoveEntity(e)
	if handle == nil {
		t.mu.Lock()
		t.travelling, t.timedOut = false, false
		t.mu.Unlock()
		return
	}

	go t.transfer(handle, source, destination, origin, pos, sourceDim, destinationDim)
}

// travelQueued moves the entity after the current transaction finishes. This is used by callers such as players that
// may touch a portal from the middle of a tick and continue running afterwards.
func (t *PortalTravelComputer) travelQueued(e Traveller, tx *world.Tx, destination *world.World) {
	source := tx.World()
	if destination == nil || destination == source {
		return
	}

	sourceDim, destinationDim := source.Dimension(), destination.Dimension()
	origin := e.Position()
	pos := translatePortalPosition(cube.PosFromVec3(origin), sourceDim, destinationDim)

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	h := e.H()
	go func() {
		var handle *world.EntityHandle
		<-source.Exec(func(tx *world.Tx) {
			// Re-open the entity in this transaction: the wrapper the travel was queued with belonged to a
			// transaction that has since finished.
			if e, ok := h.Entity(tx); ok {
				handle = tx.RemoveEntity(e)
			}
		})
		if handle == nil {
			t.mu.Lock()
			t.travelling, t.timedOut = false, false
			t.mu.Unlock()
			return
		}
		t.transfer(handle, source, destination, origin, pos, sourceDim, destinationDim)
	}()
}

// transfer adds the removed entity to the destination world at the linked portal. If no destination portal was found
// and the entity may not create one, the entity is returned to its origin in the source world instead.
func (t *PortalTravelComputer) transfer(handle *world.EntityHandle, source, destination *world.World, origin mgl64.Vec3, pos cube.Pos, sourceDim, destinationDim world.Dimension) {
	travelled := true
	<-destination.Exec(func(tx *world.Tx) {
		spawn, ok := t.destinationSpawn(tx, pos)
		if !ok {
			travelled = false
			return
		}
		if e, ok := tx.AddEntityAt(handle, spawn).(Traveller); ok {
			t.finishTravel(e, spawn, sourceDim, destinationDim)
		}
	})
	if !travelled {
		<-source.Exec(func(tx *world.Tx) {
			tx.AddEntityAt(handle, origin)
		})
	}

	t.mu.Lock()
	t.travelling = false
	t.cooldownUntil = time.Now().Add(t.Cooldown)
	if !travelled {
		// The entity is back inside the source portal: clear the arrival latch so it may retry once the
		// cooldown expires, for example after a linked portal is built.
		t.timedOut = false
	}
	t.mu.Unlock()
}

// destinationSpawn returns the position the entity should be placed at in the destination world. False is returned
// if no linked portal was found and the entity may not create one.
func (t *PortalTravelComputer) destinationSpawn(tx *world.Tx, pos cube.Pos) (mgl64.Vec3, bool) {
	if !t.CreatePortal {
		n, ok := portal.FindNetherPortal(tx, pos, portalSearchRadius)
		if !ok {
			return mgl64.Vec3{}, false
		}
		return n.Spawn().Vec3Middle(), true
	}
	if n, ok := portal.FindOrCreateNetherPortal(tx, pos, portalSearchRadius); ok {
		return n.Spawn().Vec3Middle(), true
	}
	return pos.Vec3Middle(), true
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
