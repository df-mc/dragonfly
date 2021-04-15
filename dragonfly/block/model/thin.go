package model

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Thin is a model for thin, partial blocks such as a glass pane or an iron bar. It changes its bounding box depending
// on solid faces next to it.
type Thin struct{}

// AABB ...
func (t Thin) AABB(pos cube.Pos, w *world.World) []physics.AABB {
	const offset = 0.4375

	boxes := make([]physics.AABB, 0, 5)
	mainBox := physics.NewAABB(mgl64.Vec3{offset, 0, offset}, mgl64.Vec3{1 - offset, 1, 1 - offset})

	for _, f := range cube.HorizontalFaces() {
		pos := pos.Side(f)
		block := w.Block(pos)

		// TODO(lhochbaum): Do the same check for walls as soon as they're implemented.
		if _, thin := block.Model().(Thin); thin || block.Model().FaceSolid(pos, f.Opposite(), w) {
			boxes = append(boxes, mainBox.ExtendTowards(f, offset))
		}
	}
	return append(boxes, mainBox)
}

// FaceSolid ...
func (t Thin) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	return face == cube.FaceDown
}
