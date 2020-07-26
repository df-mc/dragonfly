package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Trapdoor is a model used for trapdoors. It has no solid faces and a bounding box that changes depending on
// the direction of the trapdoor.
type Trapdoor struct {
	Facing    world.Direction
	Open, Top bool
}

// AABB ...
func (t Trapdoor) AABB(world.BlockPos, *world.World) []physics.AABB {
	if t.Open {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(t.Facing.Face()), -0.8125)}
	} else if t.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.8125}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.1875, 1})}
}

// FaceSolid ...
func (t Trapdoor) FaceSolid(world.BlockPos, world.Face, *world.World) bool {
	return false
}
