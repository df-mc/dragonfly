package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Thin is a model for thin, partial blocks such as a glass pane or an iron bar. It changes its bounding box depending
// on solid faces next to it.
type Thin struct{}

// AABB ...
func (t Thin) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	const offset = 0.4375

	boxes := make([]physics.AABB, 0, 5)
	mainBox := physics.NewAABB(mgl64.Vec3{offset, 0, offset}, mgl64.Vec3{1 - offset, 1, 1 - offset})

	for i := world.Face(2); i < 6; i++ {
		pos := pos.Side(i)
		block := w.Block(pos)

		// TODO(lhochbaum): Do the same check for walls as soon as they're implemented.
		if _, isThin := block.Model().(Thin); isThin || block.Model().FaceSolid(pos, i.Opposite(), w) {
			boxes = append(boxes, mainBox.ExtendTowards(int(i), offset))
		}
	}
	return append(boxes, mainBox)
}

// FaceSolid ...
func (t Thin) FaceSolid(_ world.BlockPos, face world.Face, _ *world.World) bool {
	return face == world.FaceDown
}
