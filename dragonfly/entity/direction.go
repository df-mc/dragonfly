package entity

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

// Facing returns the horizontal direction that an entity is facing.
func Facing(e world.Entity) world.Face {
	yaw := math.Mod(float64(e.Yaw())-90, 360)
	if yaw < 0 {
		yaw += 360
	}
	switch {
	case (yaw > 0 && yaw < 45) || (yaw > 315 && yaw < 360):
		return world.West
	case yaw > 45 && yaw < 135:
		return world.North
	case yaw > 135 && yaw < 225:
		return world.East
	case yaw > 225 && yaw < 315:
		return world.South
	}
	return 0
}

// DirectionVector returns a vector that describes the direction of the entity passed. The length of the Vec3
// returned is always 1.
func DirectionVector(e world.Entity) mgl32.Vec3 {
	yaw, pitch := float64(mgl32.DegToRad(e.Yaw())), float64(mgl32.DegToRad(e.Pitch()))
	m := math.Cos(pitch)

	return mgl32.Vec3{
		float32(-m * math.Sin(yaw)),
		float32(-math.Sin(pitch)),
		float32(m * math.Cos(yaw)),
	}.Normalize()
}
