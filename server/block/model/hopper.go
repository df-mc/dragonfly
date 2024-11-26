package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Hopper is a model used by hoppers.
type Hopper struct{}

// BBox returns a physics.BBox that spans a full block.
func (h Hopper) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	bbox := []cube.BBox{full.ExtendTowards(cube.FaceUp, -0.375)}
	for _, f := range cube.HorizontalFaces() {
		bbox = append(bbox, full.ExtendTowards(f, -0.875))
	}
	return bbox
}

// FaceSolid only returns true for the top face of the hopper.
func (Hopper) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceUp
}
