package entity

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// Rideable is an interface for entities that can be ridden.
type Rideable interface {
	// SeatPositions returns the possible seat positions for an entity in the order that they will be filled.
	SeatPositions() []mgl32.Vec3
	// Riders returns a slice entities that are currently riding an entity in the order that they were added.
	Riders() []Rider
	// AddRider adds a rider to the entity.
	AddRider(e Rider)
	// RemoveRider removes a rider from the entity.
	RemoveRider(e Rider)
	// Move moves the entity using the given vector, yaw, and pitch.
	Move(vector mgl64.Vec2, yaw, pitch float32)
}
