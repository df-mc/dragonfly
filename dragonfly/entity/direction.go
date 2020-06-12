package entity

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Facing returns the horizontal direction that an entity is facing.
func Facing(e world.Entity) world.Direction {
	yaw := math.Mod(e.Yaw()-90, 360)
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
func DirectionVector(e world.Entity) mgl64.Vec3 {
	yaw, pitch := mgl64.DegToRad(e.Yaw()), mgl64.DegToRad(e.Pitch())
	m := math.Cos(pitch)

	return mgl64.Vec3{
		-m * math.Sin(yaw),
		-math.Sin(pitch),
		m * math.Cos(yaw),
	}.Normalize()
}

// TargetBlock finds the target block of the entity passed. The block position returned will be at most
// maxDistance away from the entity. If no block can be found there, the block position returned will be
// that of an air block.
func TargetBlock(e world.Entity, maxDistance float64) world.BlockPos {
	// TODO: Implement accurate ray tracing for this.
	directionVector := DirectionVector(e)
	current := e.Position()
	if eyed, ok := e.(Eyed); ok {
		current = current.Add(mgl64.Vec3{0, eyed.EyeHeight()})
	}

	step := 0.5
	for i := 0.0; i < maxDistance; i += step {
		current = current.Add(directionVector.Mul(step))
		pos := world.BlockPosFromVec3(current)

		b := e.World().Block(pos)
		if _, ok := b.(block.Air); !ok {
			// We hit a block that isn't air.
			return pos
		}
	}
	return world.BlockPosFromVec3(current)
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
