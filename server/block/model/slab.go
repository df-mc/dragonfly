package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Slab is the model of a slab-like block, which is either a half block or a full block, depending on if the
// slab is double.
type Slab struct {
	// Double and Top specify if the Slab is a double slab and if it's in the top slot respectively. If Double is true,
	// the BBox returned is always a full block.
	Double, Top bool
}

// BBox returns either a physics.BBox spanning a full block or a half block in the top/bottom part of the block,
// depending on the Double and Top fields.
func (s Slab) BBox(cube.Pos, *world.World) []cube.BBox {
	if s.Double {
		return []cube.BBox{full}
	}
	if s.Top {
		return []cube.BBox{cube.Box(0, 0.5, 0, 1, 1, 1)}
	}
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.5, 1)}
}

// FaceSolid returns true if the Slab is double, or if the face is cube.FaceUp when the Top field is true, or if the
// face is cube.FaceDown when the Top field is false.
func (s Slab) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	if s.Double {
		return true
	} else if s.Top {
		return face == cube.FaceUp
	}
	return face == cube.FaceDown
}
