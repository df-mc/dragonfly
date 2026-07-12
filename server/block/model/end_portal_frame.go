package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EndPortalFrame is the model of a 13/16 tall, full-width end portal frame.
type EndPortalFrame struct{}

// BBox ...
func (EndPortalFrame) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.8125, 1)}
}

// FaceSolid returns true only for the down face.
func (EndPortalFrame) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceDown
}
