package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EndPortalFrame is the model of an end portal frame: a 13/16 tall full-width block with a small bump on top once an
// eye of ender has been inserted.
type EndPortalFrame struct {
	// Eye is true if an eye of ender has been inserted into the frame.
	Eye bool
}

// BBox ...
func (f EndPortalFrame) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	boxes := []cube.BBox{cube.Box(0, 0, 0, 1, 0.8125, 1)}
	if f.Eye {
		boxes = append(boxes, cube.Box(0.3125, 0.8125, 0.3125, 0.6875, 1, 0.6875))
	}
	return boxes
}

// FaceSolid returns true only for the down face.
func (EndPortalFrame) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceDown
}
