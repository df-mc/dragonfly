package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EnchantingTable is a model used by enchanting tables.
type EnchantingTable struct{}

func (EnchantingTable) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.75, 1)}
}

func (EnchantingTable) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
