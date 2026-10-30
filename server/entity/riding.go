package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Rider is an entity riding another entity.
type Rider interface {
	world.Entity

	// RidingEntity returns the entity being ridden.
	RidingEntity(tx *world.Tx) (Rideable, bool)
	// SeatIndex returns the current seat, or -1 if not riding.
	SeatIndex() int
	// MountEntity puts the rider in a seat.
	MountEntity(tx *world.Tx, rideable Rideable, seatIndex int)
	// DismountEntity removes the rider from its seat.
	DismountEntity(tx *world.Tx)
	// RidingEntityHandle returns the handle of the entity being ridden.
	RidingEntityHandle() *world.EntityHandle
	// RidingEntityController reports whether this rider controls the rideable.
	RidingEntityController() bool
	// SeatOffset returns the rider's position relative to the rideable.
	SeatOffset() (mgl64.Vec3, bool)
}

// RiderSeat holds a rider and its seat number.
type RiderSeat struct {
	Handle    *world.EntityHandle
	SeatIndex int
}

// Rideable is an entity that can have one or more riders.
type Rideable interface {
	world.Entity

	// SeatPositions returns every seat position relative to the rideable.
	SeatPositions() []mgl64.Vec3
	// NextFreeSeatIndex finds a free seat for the given click position.
	NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool)
	// ControllingRider returns the rider controlling the rideable.
	ControllingRider() *world.EntityHandle
	// ControllingSeatIndex returns the controlling seat, or -1 if there is none.
	ControllingSeatIndex() int
	// Riders returns all occupied seats.
	Riders() []RiderSeat
	// AddRider puts a rider in a seat and reports whether it succeeded.
	AddRider(rider *world.EntityHandle, seatIndex int) bool
	// RemoveRider removes a rider.
	RemoveRider(rider *world.EntityHandle)
	// MoveInput handles movement from the controlling rider.
	MoveInput(vector mgl64.Vec2, yaw, pitch float32)
}

// RiderFilter may be implemented by a Rideable that refuses some riders.
type RiderFilter interface {
	AcceptsRider(rider world.Entity) bool
}

// DismountPositioner may be implemented by a Rideable that moves riders getting off to a specific position.
type DismountPositioner interface {
	DismountPosition(rider world.Entity) mgl64.Vec3
}

// SeatRotation describes how a seat turns and restricts the rotation of its rider.
type SeatRotation struct {
	// RotateBy is the number of degrees the rider is turned by relative to the rideable.
	RotateBy float32
	// LockRotation limits the rider to looking at most LockDegrees away from the seat direction.
	LockRotation bool
	LockDegrees  float32
}

// SeatRotator may be implemented by a Rideable with seats that turn their riders.
type SeatRotator interface {
	SeatRotation(seatIndex int) SeatRotation
}

// RiderHolder may be implemented by a Rideable whose riders are not moved by knock back, pistons or portals.
type RiderHolder interface {
	HoldsRiders() bool
}

// InteractTexter may be implemented by an entity that shows a player looking at it a button, such as to sit down.
// An empty string means no button is shown.
type InteractTexter interface {
	InteractText(user world.Entity) string
}
