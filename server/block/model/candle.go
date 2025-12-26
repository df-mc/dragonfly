package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Candle is the model for candle blocks.
type Candle struct {
	Count int
}

// BBox returns the bounding box of the candle based on count.
func (c Candle) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	switch c.Count {
	case 2:
		return []cube.BBox{cube.Box(0.3125, 0, 0.4375, 0.6875, 0.375, 0.625)}
	case 3: // 3 candles
		return []cube.BBox{cube.Box(0.3125, 0, 0.375, 0.625, 0.375, 0.6875)}
	case 4: // 4 candles
		return []cube.BBox{cube.Box(0.3125, 0, 0.3125, 0.6875, 0.375, 0.625)}
	default:
		return []cube.BBox{cube.Box(0.4375, 0, 0.4375, 0.5625, 0.375, 0.5625)}
	}
}

// FaceSolid always returns false for candles.
func (Candle) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
