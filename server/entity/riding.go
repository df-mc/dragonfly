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
