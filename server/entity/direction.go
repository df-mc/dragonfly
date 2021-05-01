package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Facing returns the horizontal direction that an entity is facing.
func Facing(e world.Entity) cube.Direction {
	yaw := math.Mod(e.Yaw()-90, 360)
	if yaw < 0 {
		yaw += 360
	}
	switch {
	case (yaw > 0 && yaw < 45) || (yaw > 315 && yaw < 360):
		return cube.West
	case yaw > 45 && yaw < 135:
		return cube.North
	case yaw > 135 && yaw < 225:
		return cube.East
	case yaw > 225 && yaw < 315:
		return cube.South
	}
	return 0
}

// DirectionVector returns a vector that describes the direction of the entity passed. The length of the Vec3
// returned is always 1.
func DirectionVector(e world.Entity) mgl64.Vec3 {
	yaw, pitch := mgl64.DegToRad(e.Yaw()), mgl64.DegToRad(e.Pitch())
	m := math.Cos(pitch)

	return mgl64.Vec3{
		-m * math.Sin(yaw),
		-math.Sin(pitch),
		m * math.Cos(yaw),
	}.Normalize()
}

// Eyed represents an entity that has eyes.
type Eyed interface {
	// EyeHeight returns the offset from their base position that the eyes of an entity are found at.
	EyeHeight() float64
}

// EyePosition returns the position of the eyes of the entity if the entity implements entity.Eyed, or the
// actual position if it doesn't.
func EyePosition(e world.Entity) mgl64.Vec3 {
	pos := e.Position()
	if eyed, ok := e.(Eyed); ok {
		pos = pos.Add(mgl64.Vec3{0, eyed.EyeHeight()})
	}
	return pos
}
