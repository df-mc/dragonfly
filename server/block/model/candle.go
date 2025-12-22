package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Candle is the model for candle blocks.
type Candle struct {
	// Count is the number of candles (1-4).
	Count int
}

// BBox returns the bounding box of the candle based on count.
func (c Candle) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	// Candles have different collision boxes based on count
	switch c.Count {
	case 1:
		return []cube.BBox{cube.Box(0.4375, 0, 0.4375, 0.5625, 0.375, 0.5625)}
	case 2:
		return []cube.BBox{cube.Box(0.3125, 0, 0.3125, 0.6875, 0.375, 0.6875)}
	case 3:
		return []cube.BBox{cube.Box(0.25, 0, 0.25, 0.75, 0.375, 0.75)}
	case 4:
		return []cube.BBox{cube.Box(0.1875, 0, 0.1875, 0.8125, 0.375, 0.8125)}
	default:
		return []cube.BBox{cube.Box(0.4375, 0, 0.4375, 0.5625, 0.375, 0.5625)}
	}
}

// FaceSolid always returns false for candles.
func (Candle) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
