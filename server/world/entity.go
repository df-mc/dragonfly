package world

import (
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/state"
	"github.com/go-gl/mathgl/mgl64"
	"io"
)

// Entity represents an entity in the world, typically an object that may be moved around and can be
// interacted with by other entities.
// Viewers of a world may view an entity when near it.
type Entity interface {
	io.Closer
	// AABB returns the AABB of the entity.
	AABB() physics.AABB
	// Position returns the current position of the entity in the world.
	Position() mgl64.Vec3
	// OnGround checks if the entity is currently on the ground.
	OnGround() bool
	// World returns the current world of the entity. This is always the world that the entity can actually be
	// found in.
	World() *World
	// Rotation returns the yaw and pitch of the entity in degrees. Yaw is horizontal rotation (rotation around the
	// vertical axis, 0 when facing forward), pitch is vertical rotation (rotation around the horizontal axis, also 0
	// when facing forward).
	Rotation() (yaw, pitch float64)
	// State returns a list of entity states which the entity is currently subject to. Generally, these states
	// alter the way the entity looks.
	State() []state.State

	Velocity() mgl64.Vec3
	SetVelocity(v mgl64.Vec3)

	// Name returns a human readable name for the entity. This is not unique for an entity, but generally
	// unique for an entity type.
	Name() string

	// EncodeEntity converts the entity to its encoded representation: It returns the type of the Minecraft
	// entity, for example 'minecraft:falling_block'.
	EncodeEntity() string
}

// TickerEntity represents an entity that has a Tick method which should be called every time the entity is
// ticked every 20th of a second.
type TickerEntity interface {
	// Tick ticks the entity with the current tick passed.
	Tick(current int64)
}
