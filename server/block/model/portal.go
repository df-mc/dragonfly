package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Portal is a model used by portal blocks.
type Portal struct {
	// Axis is the axis which the portal faces.
	Axis cube.Axis
}

// BBox ...
func (p Portal) BBox(cube.Pos, *world.World) []cube.BBox {
	min, max := mgl64.Vec3{0, 0, 0.375}, mgl64.Vec3{1, 1, 0.25}
	if p.Axis == cube.Z {
		min[0], min[2], max[0], max[2] = 0.375, 0, 0.25, 1
	}
	return []cube.BBox{cube.Box(min[0], min[1], min[2], max[0], max[1], max[2])}
}

// FaceSolid ...
func (Portal) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
