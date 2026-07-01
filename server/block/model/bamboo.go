package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Bamboo is a model used by bamboo.
type Bamboo struct {
	Thick bool
}

// BBox ...
func (b Bamboo) BBox(pos cube.Pos, s world.BlockSource) []cube.BBox {
	// The stalk's box extends from the block's centre towards positive X and Z.
	size := 0.5 + 2.0/16.0
	if b.Thick {
		size = 0.5 + 3.0/16.0
	}
	// TODO: Verify the offset bounds and step count against vanilla.
	offset := randomOffset(pos, -0.25, 0.25, 16)
	return []cube.BBox{cube.Box(0.5, 0, 0.5, size, 1, size).Translate(offset)}
}

// FaceSolid ...
func (b Bamboo) FaceSolid(pos cube.Pos, face cube.Face, s world.BlockSource) bool {
	return false
}
