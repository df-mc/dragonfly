package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Fence is a model used by wooden & nether brick fence.
type Fence struct {
	Wooden bool
}

// AABB ...
func (f Fence) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	const offset = 0.375

	boxes := make([]physics.AABB, 0, 5)
	mainBox := physics.NewAABB(mgl64.Vec3{offset, 0, offset}, mgl64.Vec3{1 - offset, 1.5, 1 - offset})

	for i := world.Face(2); i < 6; i++ {
		pos := pos.Side(i)
		block := w.Block(pos)

		if fence, ok := block.Model().(Fence); (ok && fence.Wooden == f.Wooden) || block.Model().FaceSolid(pos, i, w) {
			boxes = append(boxes, mainBox.ExtendTowards(int(i), offset))
		} else if _, ok := block.Model().(FenceGate); ok {
			boxes = append(boxes, mainBox.ExtendTowards(int(i), offset))
		}
	}
	return append(boxes, mainBox)
}

// FaceSolid ...
func (f Fence) FaceSolid(_ world.BlockPos, face world.Face, _ *world.World) bool {
	return face == world.FaceDown || face == world.FaceUp
}
