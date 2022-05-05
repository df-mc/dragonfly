package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Chain is a model used by chain blocks.
type Chain struct {
	// Axis is the axis which the chain faces.
	Axis cube.Axis
}

// BBox ...
func (c Chain) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(mgl64.Vec3{0.40625, 0.40625, 0.40625}, mgl64.Vec3{0.59375, 0.59375, 0.59375}).Stretch(c.Axis, 0.40625)}
}

// FaceSolid ...
func (Chain) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
