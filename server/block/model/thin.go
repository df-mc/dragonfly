package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Thin is a model for thin, partial blocks such as a glass pane or an iron bar. It changes its bounding box depending
// on solid faces next to it.
type Thin struct{}

// BBox returns a slice of physics.BBox that depends on the blocks surrounding the Thin block. Thin blocks can connect
// to any other Thin block, wall or solid faces of other blocks.
func (t Thin) BBox(pos cube.Pos, w *world.World) []cube.BBox {
	const offset = 0.4375

	boxes := make([]cube.BBox, 0, 5)
	mainBox := cube.Box(offset, 0, offset, 1-offset, 1, 1-offset)

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

// FaceSolid returns true if the face passed is cube.FaceDown.
func (t Thin) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	return face == cube.FaceDown
}
