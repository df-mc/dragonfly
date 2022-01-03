package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Ladder is the model for a ladder block.
type Ladder struct {
	// Facing is the side opposite to the block the Ladder is currently attached to.
	Facing cube.Direction
}

// AABB returns one physics.AABB that depends on the facing direction of the Ladder.
func (l Ladder) AABB(cube.Pos, *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(l.Facing.Face(), -0.8125)}
}

// FaceSolid always returns false.
func (l Ladder) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
