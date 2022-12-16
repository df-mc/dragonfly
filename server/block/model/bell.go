package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bell represents the block model of a bell.
type Bell struct {
	// Attach represents the attachment type of the Bell.
	Attach string
	// Facing represents the direction the Bell is facing.
	Facing cube.Direction
}

// BBox ...
func (b Bell) BBox(cube.Pos, *world.World) []cube.BBox {
	if b.Attach == "standing" {
		return []cube.BBox{full.Stretch(b.Facing.Face().Axis(), -0.25).ExtendTowards(cube.FaceUp, -0.1875)}
	}
	if b.Attach == "hanging" {
		return []cube.BBox{full.GrowVec3(mgl64.Vec3{-0.25, 0, -0.25}).ExtendTowards(cube.FaceDown, -0.25)}
	}

	box := full.Stretch(b.Facing.RotateLeft().Face().Axis(), -0.25).
		ExtendTowards(cube.FaceUp, -0.0625).
		ExtendTowards(cube.FaceDown, -0.25)
	if b.Attach == "side" {
		return []cube.BBox{box.ExtendTowards(b.Facing.Face(), -0.1875)}
	}
	return []cube.BBox{box}
}

// FaceSolid ...
func (Bell) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
