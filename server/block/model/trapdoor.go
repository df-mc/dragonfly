package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
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

// BBox returns a physics.BBox that depends on the facing direction of the Trapdoor and whether it is open and in the
// top part of the block.
func (t Trapdoor) BBox(cube.Pos, *world.World) []cube.BBox {
	if t.Open {
		return []cube.BBox{full.ExtendTowards(t.Facing.Face(), -0.8125)}
	} else if t.Top {
		return []cube.BBox{cube.Box(0, 0.8125, 0, 1, 1, 1)}
	}
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.1875, 1)}
}

// FaceSolid always returns false.
func (t Trapdoor) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
