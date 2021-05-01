package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Trapdoor is a model used for trapdoors. It has no solid faces and a bounding box that changes depending on
// the direction of the trapdoor.
type Trapdoor struct {
	Facing    cube.Direction
	Open, Top bool
}

// AABB ...
func (t Trapdoor) AABB(cube.Pos, *world.World) []physics.AABB {
	if t.Open {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(t.Facing.Face(), -0.8125)}
	} else if t.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.8125}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.1875, 1})}
}

// FaceSolid ...
func (t Trapdoor) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
