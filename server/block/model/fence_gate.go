package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// FenceGate is a model used by fence gates. The model is completely zero-ed when the FenceGate is opened.
type FenceGate struct {
	// Facing is the facing direction of the FenceGate. A fence gate can only be opened in this direction or the
	// direction opposite to it.
	Facing cube.Direction
	// Open specifies if the FenceGate is open. In this case, AABB returns an empty slice.
	Open bool
}

// AABB returns up to one physics.AABB depending on the facing direction of the FenceGate and whether it is open.
func (f FenceGate) AABB(cube.Pos, *world.World) []physics.AABB {
	if f.Open {
		return nil
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1.5, 1}).Stretch(f.Facing.Face().Axis(), -0.375)}
}

// FaceSolid always returns false.
func (f FenceGate) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
