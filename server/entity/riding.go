package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Rider is an entity that can ride another entity.
//
// Entity values obtained from a transaction are only valid for that
// transaction. RidingEntity therefore resolves the stored relationship using
// the transaction supplied by the caller, while the relationship itself is
// kept by an entity handle.
type Rider interface {
	world.Entity

	// RidingEntity returns the rideable in tx, if the rider is mounted and the
	// rideable is still in the same world.
	RidingEntity(tx *world.Tx) (Rideable, bool)
	// SeatIndex returns the rider's seat index, or -1 when the rider is not
	// mounted.
	SeatIndex() int
	// MountEntity mounts the rider at seatIndex when that seat is valid and
	// available.
	MountEntity(tx *world.Tx, rideable Rideable, seatIndex int)
	// DismountEntity dismounts the rider. Calling it while unmounted is safe.
	DismountEntity(tx *world.Tx)
	// RidingEntityHandle returns the stable handle of the current rideable, or
	// nil when the rider is not mounted.
	RidingEntityHandle() *world.EntityHandle
	// RidingEntityController reports the controller state last synchronised from
	// the rideable. It is used when replaying links after this rider becomes
	// visible to a client.
	RidingEntityController() bool
	// SeatOffset returns the value copied from the rideable's current seat
	// geometry for metadata encoding.
	SeatOffset() (mgl64.Vec3, bool)
}

// RiderSeat identifies a registered rider without extending the lifetime of
// its transaction-scoped entity value.
type RiderSeat struct {
	Handle    *world.EntityHandle
	SeatIndex int
}

// Rideable is an entity that can have one or more riders.
type Rideable interface {
	world.Entity

	// SeatPositions returns seat offsets relative to the rideable's position.
	// The index in the returned slice is the seat index.
	SeatPositions() []mgl64.Vec3
	// NextFreeSeatIndex returns the preferred free seat for a click position,
	// or (-1, false) when no seat is available.
	NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool)
	// ControllingRider returns the stable handle of the rider that currently
	// controls the rideable, or nil. A rideable may choose its controller
	// independently of seat order.
	ControllingRider() *world.EntityHandle
	// ControllingSeatIndex returns the seat index currently advertised as
	// controlling, or -1 when there is no controller. It must be the
	// registration associated with ControllingRider.
	ControllingSeatIndex() int
	// Riders returns handle-backed seat assignments. The rideable owns these
	// registrations and must not retain entity values obtained from a
	// transaction. While operating in a transaction, it must remove a
	// registration whose handle cannot resolve in that transaction.
	Riders() []RiderSeat
	// AddRider atomically registers rider at seatIndex and returns whether the
	// assignment was accepted. Re-registering the same handle changes its seat.
	// A false result leaves the existing assignment unchanged.
	AddRider(rider *world.EntityHandle, seatIndex int) bool
	// RemoveRider unregisters rider. It is safe when rider is not registered.
	RemoveRider(rider *world.EntityHandle)
	// MoveInput forwards input from the controlling rider.
	MoveInput(vector mgl64.Vec2, yaw, pitch float32)
}
