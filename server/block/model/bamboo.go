package model

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bamboo is a model used by bamboo.
type Bamboo struct {
	Thick bool
}

// BBox ...
func (b Bamboo) BBox(pos cube.Pos, s world.BlockSource) []cube.BBox {
	width := 2.0
	if b.Thick {
		width = 3.0
	}
	inset := 1.0 - (width / 16.0)

	seed := b.OffsetSeed(pos)
	offsetX := float64((seed%12)+1) / 16.0
	offsetZ := float64(((seed>>8)%12)+1) / 16.0

	return []cube.BBox{full.ExtendTowards(cube.FaceSouth, -inset).ExtendTowards(cube.FaceEast, -inset).GrowVec3(mgl64.Vec3{offsetX, 0, offsetZ})}
}

// OffsetSeed returns a seed based on the position of the bamboo to offset its model.
func (b Bamboo) OffsetSeed(pos cube.Pos) int {
	p1 := pos.Z() * 116129781
	p2 := pos.X() * 3129871
	xord := (p1 ^ p2) ^ pos.Y()
	return (((xord * 42317861) + 11) * xord) & math.MaxUint32
}

// FaceSolid ...
func (b Bamboo) FaceSolid(pos cube.Pos, face cube.Face, s world.BlockSource) bool {
	return false
}
