package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

const (
	wallConnectionTypeNone  = "none"
	wallConnectionTypeShort = "short"
	wallConnectionTypeTall  = "tall"
)

// Wall is a model used by all wall types.
type Wall struct {
	// NorthConnection is the type of connection for the north direction. This can be any of the constants above.
	NorthConnection string
	// EastConnection is the type of connection for the east direction. This can be any of the constants above.
	EastConnection string
	// SouthConnection is the type of connection for the south direction. This can be any of the constants above.
	SouthConnection string
	// WestConnection is the type of connection for the west direction. This can be any of the constants above.
	WestConnection string
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
	if w.NorthConnection != wallConnectionTypeNone {
		boxes = append(boxes, cube.Box(0.25, 0, 0, 0.75, w.heightFromConnection(w.NorthConnection), 0.25))
	}
	if w.EastConnection != wallConnectionTypeNone {
		boxes = append(boxes, cube.Box(0.75, 0, 0.25, 1, w.heightFromConnection(w.EastConnection), 0.75))
	}
	if w.SouthConnection != wallConnectionTypeNone {
		boxes = append(boxes, cube.Box(0.25, 0, 0.75, 0.75, w.heightFromConnection(w.SouthConnection), 1))
	}
	if w.WestConnection != wallConnectionTypeNone {
		boxes = append(boxes, cube.Box(0, 0, 0.25, 0.25, w.heightFromConnection(w.WestConnection), 0.75))
	}
	return boxes
}

// FaceSolid ...
func (w Wall) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

// heightFromConnection calculates the height of a connection based on the provided connection type.
func (w Wall) heightFromConnection(connection string) float64 {
	if connection == wallConnectionTypeTall {
		return 1
	}
	return 0.75
}
