package world

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/item"
	"github.com/go-gl/mathgl/mgl32"
	"io"
)

// Entity represents an entity in the world, typically an object that may be moved around and can be
// interacted with by other entities.
// Viewers of a world may view an entity when near it.
type Entity interface {
	io.Closer
	// Pos returns the current position of the entity in the world.
	Position() mgl32.Vec3
	// World returns the current world of the entity. This is always the world that the entity can actually be
	// found in.
	World() *World
	// Yaw returns the yaw of the entity. This is horizontal rotation (rotation around the vertical axis), and
	// is 0 when the entity faces forward.
	Yaw() float32
	// Pitch returns the pitch of the entity. This is vertical rotation (rotation around the horizontal axis),
	// and is 0 when the entity faces forward.
	Pitch() float32

	setPosition(new mgl32.Vec3)
	setYaw(new float32)
	setPitch(new float32)
	setWorld(new *World)
}

// CarryingEntity represents an entity that is able to carry items with it.
type CarryingEntity interface {
	Entity
	// HeldItems returns the items currently held by the entity. Viewers of the entity will be able to see
	// these items.
	HeldItems() (mainHand, offHand item.Stack)
}
