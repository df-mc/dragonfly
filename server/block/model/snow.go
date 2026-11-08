package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Snow is the model of a snow layer. The height of its bounding box scales with the number of layers the block
// has, ranging from a single passable layer to a near-full block.
type Snow struct {
	// Layers is the number of snow layers the block has, ranging from 1 to 8.
	Layers int
}

// BBox returns a flat box whose height scales with the number of layers. A single layer returns no box at all, so
// that entities walk over it freely as they do in vanilla.
func (s Snow) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	if s.Layers <= 1 {
		return nil
	}
	return []cube.BBox{cube.Box(0, 0, 0, 1, float64(s.Layers-1)*0.125, 1)}
}

// FaceSolid returns true only for the upward face of a full eight-layer block, the only state to which other
// blocks may be attached.
func (s Snow) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return s.Layers >= 8 && face == cube.FaceUp
}
