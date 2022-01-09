package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Skull is a model used by skull blocks.
type Skull struct {
	// Direction is the direction the skull is facing.
	Direction cube.Face
}

// AABB ...
func (s Skull) AABB(cube.Pos, *world.World) []physics.AABB {
	aabb := physics.NewAABB(mgl64.Vec3{0.25, 0, 0.25}, mgl64.Vec3{0.75, 0.5, 0.75})
	if s.Direction.Axis() == cube.Y {
		return []physics.AABB{aabb}
	}
	return []physics.AABB{aabb.TranslateTowards(s.Direction.Opposite(), 0.25).TranslateTowards(cube.FaceUp, 0.25)}
}

// FaceSolid ...
func (Skull) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
