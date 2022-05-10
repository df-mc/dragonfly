package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Fence is a model used by fences of any type. It can attach to blocks with solid faces and other fences of the same
// type and has a model height just slightly over 1.
type Fence struct {
	// Wood specifies if the Fence is made from wood. This field is used to check if two fences are able to attach to
	// each other.
	Wood bool
}

// BBox returns multiple physics.BBox depending on how many connections it has with the surrounding blocks.
func (f Fence) BBox(pos cube.Pos, w *world.World) []cube.BBox {
	const offset = 0.375

	boxes := make([]cube.BBox, 0, 5)
	mainBox := cube.Box(offset, 0, offset, 1-offset, 1.5, 1-offset)

	for i := cube.Face(2); i < 6; i++ {
		pos := pos.Side(i)
		block := w.Block(pos)

		if fence, ok := block.Model().(Fence); (ok && fence.Wood == f.Wood) || block.Model().FaceSolid(pos, i, w) {
			boxes = append(boxes, mainBox.ExtendTowards(i, offset))
		} else if _, ok := block.Model().(FenceGate); ok {
			boxes = append(boxes, mainBox.ExtendTowards(i, offset))
		}
	}
	return append(boxes, mainBox)
}

// FaceSolid returns true if the face is cube.FaceDown or cube.FaceUp.
func (f Fence) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	return face == cube.FaceDown || face == cube.FaceUp
}
