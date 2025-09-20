package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// BrewingStand is a model used by brewing stands.
type BrewingStand struct{}

func (b BrewingStand) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{
		full.ExtendTowards(cube.FaceDown, 0.875),
		full.Stretch(cube.X, -0.4375).Stretch(cube.Z, -0.4375).ExtendTowards(cube.FaceDown, 0.125),
	}
}

func (b BrewingStand) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
