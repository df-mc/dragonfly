package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

/*
Rider / Rideable — detailed developer documentation

Overview
--------
This file declares the Rider and Rideable interfaces which form the "mounting" subsystem used by entities
that can sit on (Rider) or be sat upon (Rideable). The documentation below explains the exact contract,
expected invariants, typical usage flows and implementation notes for authors who implement either
interface or interact with objects that implement them.

Key concepts
------------
- Seat index: an integer referring to a seat returned by a Rideable's SeatPositions(). Seat indices are
  zero-based. A Rider must use -1 to indicate "not seated".
- Seat position: an mgl64.Vec3 offset returned by SeatPositions(); offsets are relative to the rideable
  entity origin. Callers that need world-space seat coordinates must add the rideable entity's world
  position to this offset.
- Controlling rider: a Rideable may expose a controlling rider (via ControllingRider()) that is allowed
  to provide movement input (MoveInput()). Immobile rideables should still implement MoveInput — it may
  be a no-op, but storing input state may be useful for aiming/animation.

Interfaces (contract)
---------------------
Rider:
- RidingEntity() Rideable
  - Returns the Rideable currently mounted, or nil when not mounted.
- SeatIndex() int
  - Returns the current seat index, or -1 when not mounted.
- MountEntity(rideable Rideable, seatIndex int)
  - Mount the rider onto the given rideable at the given seat index. Implementations MUST validate the
    provided seatIndex before using it, and maintain invariants described below.
- DismountEntity()
  - Dismount from the current rideable. Calling this when not mounted must be a no-op.

Rideable:
- SeatPositions() []mgl64.Vec3
  - Returns the seat offsets relative to the rideable origin. The returned slice may be a copy or a
    read-only view; callers must not mutate it.
- NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool)
  - Given a click position (in world-space), returns the preferred free seat index and true, or -1 and
    false if none are available.
- ControllingRider() Rider
  - Returns the rider that currently controls the rideable (may be nil).
- Riders() []Rider
  - Returns a slice of current riders (may be nil or empty if no riders are tracked).
- AddRider(rider Rider)
  - Register a rider on the rideable. Implementations that track riders should use rider.SeatIndex()
    to determine placement; implementations that don't track riders may choose to be no-ops.
- RemoveRider(rider Rider)
  - Unregister a rider. Must be safe to call even if the rider is not present.
- MoveInput(vector mgl64.Vec2, yaw, pitch float32)
  - Apply movement/aim input from the controlling rider. For immobile rideables this may be a no-op,
    but it should at least store state if other systems rely on it.

Invariants and required behaviour
---------------------------------
- Riders must maintain the invariant: if RidingEntity() == nil then SeatIndex() == -1. Conversely, a
  non-nil rideable with SeatIndex() >= 0 must be within the bounds of rideable.SeatPositions().
- Always bounds-check SeatIndex() before indexing into SeatPositions().
- MountEntity should call DismountEntity first if already mounted (to ensure consistent notifications
  and state transitions).
- When both sides (rider and rideable) track state, keep them consistent: MountEntity should update the
  rider-side state then call rideable.AddRider(r), and DismountEntity should call rideable.RemoveRider(r)
  before clearing rider-side state, or follow a project-wide convention. Document which side owns "the
  source of truth" for rider lists in your implementation.

Viewer and state notifications
------------------------------
The project notifies viewers when mount/dismount/seat-change events occur (for example via
ViewEntityMount, ViewEntityDismount and ViewEntityState). Implementations should:
- Trigger viewer notifications consistently when mounting, dismounting and changing seats.
- Preserve notification ordering used elsewhere in the codebase (typically: update internal state,
  call updateState() or a similar method, then send mount/dismount packets to viewers).

Concurrency
-----------
- If rideable or rider state is accessed from multiple goroutines, protect mutable state with a mutex or
  ensure all modifications happen on the main server tick.
- Methods that return slices (SeatPositions, Riders) may return copies to avoid races; callers should not
  mutate returned slices.

Usage patterns (recommended)
----------------------------
Mount flow (server tick):
1. Identify the rideable the player clicked and compute clickPos (world-space).
2. idx, ok := rideable.NextFreeSeatIndex(clickPos)
3. If ok: call rider.MountEntity(rideable, idx). If the rideable tracks riders and the project's
   convention requires it, call rideable.AddRider(rider) as well (some implementations expect rider-side
   MountEntity to call AddRider internally — follow the codebase convention).
4. Notify viewers using project's viewer API.

Dismount flow:
1. Call rider.DismountEntity(). If the rideable tracks riders and the project's convention requires it,
   ensure rideable.RemoveRider(rider) is called (either by the rider's DismountEntity implementation or by
   the caller).

Seat change:
- Prefer calling a dedicated ChangeSeat method on the rider if available. Validate the new seat index
  against rideable.SeatPositions() before applying.

TWEntity behaviour reference
----------------------------
The repository contains a concrete TWEntity implementation used as an example/reference. Below is the
complete TWEntity implementation (included here for reference in the public docs).

package entity

import (
	"sync"

	"github.com/go-gl/mathgl/mgl64"
)

// TWEntity implements both Rideable and Rider completely. It tracks seats, riders, controlling rider,
// and preserves viewer notification behaviour. This reference is included in the docs so implementers
// have a complete example to follow.
type TWEntity struct {
	mu sync.Mutex

	// Rideable-side state
	seats       []mgl64.Vec3      // seat offsets relative to entity origin
	riders      map[int]Rider     // seatIndex -> Rider
	controlling Rider             // current controlling rider, if any
	position    mgl64.Vec3        // world position of the entity (used for seat world pos)

	// Rider-side state
	rideable  Rideable
	seatIndex int // -1 when not seated

	// Misc
	lastMoveInput struct {
		vector mgl64.Vec2
		yaw    float32
		pitch  float32
	}
}

// NewTWEntity returns a TWEntity initialised with a single seat {0,2,0}.
func NewTWEntity() *TWEntity {
	return &TWEntity{
		seats:     []mgl64.Vec3{{0, 2, 0}},
		riders:    make(map[int]Rider),
		seatIndex: -1,
	}
}

// --- Rideable methods ---

// SeatPositions returns a copy of seat offsets relative to the entity origin.
func (e *TWEntity) SeatPositions() []mgl64.Vec3 {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]mgl64.Vec3, len(e.seats))
	copy(out, e.seats)
	return out
}

// NextFreeSeatIndex returns the nearest free seat index to clickPos (in world space).
// If no free seat is available, returns (-1, false).
func (e *TWEntity) NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.seats) == 0 {
		return -1, false
	}
	bestIdx := -1
	var bestDist float64
	for i, off := range e.seats {
		if _, occupied := e.riders[i]; occupied {
			continue
		}
		seatWorld := e.position.Add(off)
		d := seatWorld.Sub(clickPos).Len()
		if bestIdx == -1 || d < bestDist {
			bestIdx = i
			bestDist = d
		}
	}
	if bestIdx == -1 {
		return -1, false
	}
	return bestIdx, true
}

// ControllingRider returns the rider currently controlling this entity, if any.
func (e *TWEntity) ControllingRider() Rider {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.controlling
}

// Riders returns a slice of all current riders.
func (e *TWEntity) Riders() []Rider {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Rider, 0, len(e.riders))
	for _, r := range e.riders {
		out = append(out, r)
	}
	return out
}

// AddRider registers a rider into the rideable's internal tracking. It expects the rider's SeatIndex()
// to be already set to the intended seat.
func (e *TWEntity) AddRider(r Rider) {
	if r == nil {
		return
	}
	idx := r.SeatIndex()
	e.mu.Lock()
	defer e.mu.Unlock()
	// Validate seat index
	if idx < 0 || idx >= len(e.seats) {
		return
	}
	// Prevent double-assign
	if cur, ok := e.riders[idx]; ok {
		if cur == r {
			return
		}
		// seat occupied by another rider; do not overwrite
		return
	}
	e.riders[idx] = r
	if e.controlling == nil {
		e.controlling = r
	}
}

// RemoveRider removes the rider from internal tracking if present.
func (e *TWEntity) RemoveRider(r Rider) {
	if r == nil {
		return
	}
	idx := r.SeatIndex()
	e.mu.Lock()
	defer e.mu.Unlock()
	if idx < 0 {
		// try to remove any matching rider by reference
		for i, rr := range e.riders {
			if rr == r {
				delete(e.riders, i)
				if e.controlling == r {
					e.controlling = nil
					for _, other := range e.riders {
						e.controlling = other
						break
					}
				}
				return
			}
		}
		return
	}
	if e.riders[idx] == r {
		delete(e.riders, idx)
		if e.controlling == r {
			e.controlling = nil
			for _, other := range e.riders {
				e.controlling = other
				break
			}
		}
	}
}

// MoveInput stores the last received input. For immobile entities we still record
// the aiming/input state so that viewers or other systems can react.
func (e *TWEntity) MoveInput(vector mgl64.Vec2, yaw, pitch float32) {
	e.mu.Lock()
	e.lastMoveInput.vector = vector
	e.lastMoveInput.yaw = yaw
	e.lastMoveInput.pitch = pitch
	e.mu.Unlock()
	// Notify viewers that the entity state changed (e.g. aiming changed).
	e.updateState()
}

// --- Rider methods ---

// RidingEntity returns the rideable the rider is currently mounted on (or nil).
func (e *TWEntity) RidingEntity() Rideable {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.rideable
}

// SeatIndex returns the current seat index, or -1 when not mounted.
func (e *TWEntity) SeatIndex() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.seatIndex
}

// ChangeSeat moves the rider to another seat on the same rideable.
// Properly updates rideable tracking.
func (e *TWEntity) ChangeSeat(seatIndex int) {
	e.mu.Lock()
	r := e.rideable
	e.mu.Unlock()
	if r == nil {
		return
	}
	seats := r.SeatPositions()
	if seatIndex < 0 || seatIndex >= len(seats) {
		return
	}
	// Remove from previous seat on rideable, update index, add to new seat
	r.RemoveRider(e) // safe even if not present
	e.mu.Lock()
	e.seatIndex = seatIndex
	e.mu.Unlock()
	r.AddRider(e)

	e.updateState()
	for _, v := range e.Tx().Viewers(e.Position()) {
		v.ViewEntityMount(e, r, e.seatIndex == 0)
	}
}

// SeatPosition returns the seat offset relative to the rideable origin and whether seated.
func (e *TWEntity) SeatPosition() (mgl64.Vec3, bool) {
	e.mu.Lock()
	r := e.rideable
	idx := e.seatIndex
	e.mu.Unlock()
	if r == nil || idx < 0 {
		return mgl64.Vec3{}, false
	}
	seats := r.SeatPositions()
	if idx < 0 || idx >= len(seats) {
		return mgl64.Vec3{}, false
	}
	return seats[idx], true
}

// MountEntity mounts this entity onto a rideable and registers it on the rideable.
func (e *TWEntity) MountEntity(rideable Rideable, seatIndex int) {
	if rideable == nil {
		return
	}
	// If already mounted, dismount first.
	e.DismountEntity()

	// Validate seat index against the rideable's seats.
	seats := rideable.SeatPositions()
	if seatIndex < 0 || seatIndex >= len(seats) {
		return
	}

	// Set rider-side state first so rideable.AddRider can trust SeatIndex().
	e.mu.Lock()
	e.rideable = rideable
	e.seatIndex = seatIndex
	e.mu.Unlock()

	// Register on rideable side.
	rideable.AddRider(e)

	e.updateState()
	for _, v := range e.Tx().Viewers(e.Position()) {
		v.ViewEntityMount(e, rideable, seatIndex == 0)
	}
}

// DismountEntity removes this rider from its rideable and updates state.
func (e *TWEntity) DismountEntity() {
	e.mu.Lock()
	r := e.rideable
	e.mu.Unlock()
	if r == nil {
		return
	}

	// Remove from rideable's tracking first.
	r.RemoveRider(e)

	// Clear rider state.
	e.mu.Lock()
	e.rideable = nil
	e.seatIndex = -1
	e.mu.Unlock()

	e.updateState()
	for _, v := range e.Tx().Viewers(e.Position()) {
		v.ViewEntityDismount(e, r)
	}
}

// updateState notifies viewers about the entity state (keeps your original behaviour).
func (e *TWEntity) updateState() {
	for _, v := range e.Tx().Viewers(e.Position()) {
		v.ViewEntityState(e)
	}
}

Implementer's checklist
-----------------------
- [ ] Ensure SeatPositions returns offsets (relative to entity origin).
- [ ] Ensure MountEntity validates seatIndex and preserves invariants.
- [ ] Ensure DismountEntity is safe to call when not mounted.
- [ ] Ensure AddRider and RemoveRider leave the rideable's internal state consistent.
- [ ] Protect mutable shared state when accessed concurrently.
- [ ] Emit viewer notifications in the same places other entities do in this codebase (updateState,
      ViewEntityMount/ViewEntityDismount/ViewEntityState).

Common pitfalls
---------------
- Indexing SeatPositions without checking SeatIndex can cause panics.
- Forgetting to clear seatIndex on dismount (breaking the Rider invariant).
- Inconsistent updates between rider and rideable when both track riders — always update both sides or
  centralize ownership.

Quick examples (pseudo / reference snippets)
-------------------------------------------
The following are concise reference snippets. They are not full implementations — see concrete types in
this package for working examples.

Mount helper (recommended flow):

	idx, ok := rideable.NextFreeSeatIndex(clickPos)
	if ok {
		r.MountEntity(rideable, idx)
		// if the project's convention requires: rideable.AddRider(r)
	}

Seat offset lookup from a rider (returns offset relative to rideable origin):

	if off, ok := r.(interface{ SeatPosition() (mgl64.Vec3, bool) }); ok {
		offset, _ := off.SeatPosition()
		_ = offset // use offset relative to rideable origin
	}
*/

// Rider is an interface for entities that can ride other entities.
type Rider interface {
	world.Entity
	// RidingEntity returns the entity that the rider is currently sitting on.
	RidingEntity(tx *world.Tx) (Rideable, bool)
	// SeatIndex returns the position of where the rider is sitting.
	SeatIndex() int
	// MountEntity mounts the Rider to an entity if the entity is Rideable and if there is a seat available.
	MountEntity(tx *world.Tx, rideable Rideable, seatIndex int)
	// DismountEntity dismounts the rider from the entity they are currently riding.
	DismountEntity(tx *world.Tx)
}

// Rideable is an interface for entities that can be ridden.
type Rideable interface {
	world.Entity
	// SeatPositions returns a map of seat indices to their positions relative to the entity's position.
	SeatPositions() []mgl64.Vec3
	// NextFreeSeatIndex returns the index of the next free seat and whether a free seat was found.
	NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool)
	// ControllingRider returns the rider that is controlling the entity, if any.
	ControllingRider() Rider
	// Riders returns a slice of all riders currently riding the entity.
	Riders() []Rider
	// AddRider adds a rider to the entity.
	AddRider(rider Rider)
	// RemoveRider removes a rider from the entity.
	RemoveRider(rider Rider)
	// MoveInput moves the entity based on input from the controlling rider.
	MoveInput(vector mgl64.Vec2, yaw, pitch float32)
}
