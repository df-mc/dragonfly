package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Shulker is the model of a shulker box. The bounding box grows along the
// facing axis as the lid opens.
type Shulker struct {
	// Facing is the direction that the lid opens towards.
	Facing cube.Face
	// Progress is the lid animation progress, ranging from 0 (closed) to 10 (fully open).
	Progress int32
}

// BBox returns a single bounding box that extends outward along Facing and
// inward along its opposite as the lid opens.
func (s Shulker) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	peak := ShulkerPhysicalPeak(s.Progress)
	return []cube.BBox{full.ExtendTowards(s.Facing, peak).ExtendTowards(s.Facing.Opposite(), -peak)}
}

// ShulkerPhysicalPeak returns the lid extension along the facing axis for a
// given Progress in [0, 10]. The curve eases out cubically so the lid moves
// quickly and settles.
func ShulkerPhysicalPeak(progress int32) float64 {
	t := float64(progress) / 10.0
	return (1.0 - (1.0-t)*(1.0-t)*(1.0-t)) * 0.5
}

// FaceSolid always returns false.
func (Shulker) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
