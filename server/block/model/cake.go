package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Cake is a model used by cake blocks.
type Cake struct {
	// Bites is the amount of bites that were taken from the cake. A cake can have up to 7 bites taken from it, before
	// being consumed entirely.
	Bites int
}

// BBox returns an BBox with a size that depends on the amount of bites taken.
func (c Cake) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0.0625, 0, 0.0625, 0.9375, 0.5, 0.9375).
		ExtendTowards(cube.FaceWest, -(float64(c.Bites) / 8))}
}

// FaceSolid always returns false.
func (c Cake) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
