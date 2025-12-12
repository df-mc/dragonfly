package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Shulker is the model of a shulker. It depends on the opening/closing progress of the shulker block.
type Shulker struct {
	// Facing is the face that the shulker faces.
	Facing cube.Face
	// Progress is the opening/closing progress of the shulker.
	Progress int32
}

// BBox returns a BBox that depends on the opening/closing progress of the shulker.
func (s Shulker) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	peak := physicalPeak(s.Progress)
	// Adds peak to the top and subtracts peak from the bottom.  (according to BDS)
	bbox := full
	bbox.ExtendTowards(s.Facing, peak).ExtendTowards(s.Facing.Opposite(), -peak)

	return []cube.BBox{bbox}
}

// physicalPeak returns the peak of which the shulker reaches in its current progress
func physicalPeak(progress int32) float64 {
	fp := float64(progress) / 10.0
	openness := 1.0 - fp
	return (1.0 - openness*openness*openness) * 0.5
}

// FaceSolid always returns false.
func (Shulker) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
