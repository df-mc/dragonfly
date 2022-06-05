package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type Wall struct {
	NorthConnection bool
	EastConnection  bool
	SouthConnection bool
	WestConnection  bool
	Post            bool
}

// BBox ...
func (w Wall) BBox(cube.Pos, *world.World) []cube.BBox {
	height := 0.8125
	if w.Post {
		height = 1
	}
	boxes := []cube.BBox{cube.Box(0.25, 0, 0.25, 0.75, height, 0.75)}
	if w.NorthConnection {
		boxes = append(boxes, cube.Box(0.25, 0, 0, 0.75, height, 0.25)) // TODO: Connection height
	}
	if w.EastConnection {
		boxes = append(boxes, cube.Box(0.75, 0, 0.25, 1, height, 0.75)) // TODO: Connection height
	}
	if w.SouthConnection {
		boxes = append(boxes, cube.Box(0.25, 0, 0.75, 0.75, height, 1)) // TODO: Connection height
	}
	if w.WestConnection {
		boxes = append(boxes, cube.Box(0, 0, 0.25, 0.25, height, 0.75)) // TODO: Connection height
	}
	return boxes
}

// FaceSolid ...
func (w Wall) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}
