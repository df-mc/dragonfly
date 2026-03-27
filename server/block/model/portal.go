package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Portal is a model used by Nether portal blocks.
type Portal struct {
	// Axis is the axis normal to the portal plane.
	Axis cube.Axis
}

// BBox returns the thin collision-free portal plane.
func (p Portal) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	min, max := mgl64.Vec3{0, 0, 0.375}, mgl64.Vec3{1, 1, 0.625}
	if p.Axis == cube.Z {
		min[0], min[2], max[0], max[2] = 0.375, 0, 0.625, 1
	}
	return []cube.BBox{cube.Box(min[0], min[1], min[2], max[0], max[1], max[2])}
}

// FaceSolid always returns false.
func (Portal) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
