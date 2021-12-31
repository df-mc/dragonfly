package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Lantern is a model for the lantern block. It can be placed on the ground or hanging from the ceiling.
type Lantern struct {
	// Hanging specifies if the lantern is hanging from a block or if it's placed on the ground.
	Hanging bool
}

// AABB returns a physics.AABB attached to either the ceiling or to the ground.
func (l Lantern) AABB(cube.Pos, *world.World) []physics.AABB {
	if l.Hanging {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0.3125, 0.125, 0.3125}, mgl64.Vec3{0.6875, 0.625, 0.6875})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{0.3125, 0, 0.3125}, mgl64.Vec3{0.6875, 0.5, 0.6875})}
}

// FaceSolid always returns false.
func (l Lantern) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
