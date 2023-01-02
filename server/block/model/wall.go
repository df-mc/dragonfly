package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Wall is a model used by all wall types.
type Wall struct {
	// NorthConnection is the height of the connection for the north direction.
	NorthConnection float64
	// EastConnection is the height of the connection for the east direction.
	EastConnection float64
	// SouthConnection is the height of the connection for the south direction.
	SouthConnection float64
	// WestConnection is the height of the connection for the west direction.
	WestConnection float64
	// Post is if the wall is the full height of a block or not.
	Post bool
}

// BBox ...
func (w Wall) BBox(cube.Pos, *world.World) []cube.BBox {
	postHeight := 0.8125
	if w.Post {
		postHeight = 1
	}
	boxes := []cube.BBox{cube.Box(0.25, 0, 0.25, 0.75, postHeight, 0.75)}
	if w.NorthConnection > 0 {
		boxes = append(boxes, cube.Box(0.25, 0, 0.75, 0.75, w.SouthConnection, 1))
	}
	if w.EastConnection > 0 {
		boxes = append(boxes, cube.Box(0, 0, 0.25, 0.25, w.WestConnection, 0.75))
	}
	if w.SouthConnection > 0 {
		boxes = append(boxes, cube.Box(0.25, 0, 0, 0.75, w.NorthConnection, 0.25))
	}
	if w.WestConnection > 0 {
		boxes = append(boxes, cube.Box(0.75, 0, 0.25, 1, w.EastConnection, 0.75))
	}
	return boxes
}

// FaceSolid returns true if the face is in the Y axis.
func (w Wall) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	return face.Axis() == cube.Y
}
