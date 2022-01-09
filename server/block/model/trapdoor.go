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
	// Facing is the facing direction of the Trapdoor. In addition to the texture, it influences the direction in which
	// the Trapdoor is opened.
	Facing cube.Direction
	// Open and Top specify if the Trapdoor is opened and if it's in the top or bottom part of a block respectively.
	Open, Top bool
}

// AABB returns a physics.AABB that depends on the facing direction of the Trapdoor and whether it is open and in the
// top part of the block.
func (t Trapdoor) AABB(cube.Pos, *world.World) []physics.AABB {
	if t.Open {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(t.Facing.Face(), -0.8125)}
	} else if t.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.8125}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.1875, 1})}
}

// FaceSolid always returns false.
func (t Trapdoor) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
