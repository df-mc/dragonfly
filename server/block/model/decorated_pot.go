package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// DecoratedPot is the model for a DecoratedPot. It is just barely not a full block, having a slightly reduced width and
// depth.
type DecoratedPot struct{}

// BBox returns a physics.BBox that is slightly smaller than a full block.
func (DecoratedPot) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0.025, 0, 0.025, 0.975, 1, 0.975)}
}

// FaceSolid always returns true for the top and bottom face, and false for all other faces.
func (DecoratedPot) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face.Axis() == cube.Y
}
