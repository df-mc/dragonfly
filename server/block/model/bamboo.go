package model

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/mcrandom"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bamboo is a model used by bamboo.
type Bamboo struct {
	Thick bool
}

// BBox ...
func (b Bamboo) BBox(pos cube.Pos, s world.BlockSource) []cube.BBox {
	pixels := 2.0
	if b.Thick {
		pixels = 3.0
	}
	radius := (pixels / 16.0) / 2.0
	return []cube.BBox{cube.Box(0.5-radius, 0, 0.5-radius, 0.5+radius, 1, 0.5+radius).Translate(b.randomlyModifyPosition(pos))}
}

// positionHash ...
func (Bamboo) positionHash(x, z int) uint64 {
	ux := uint64(uint32(x))
	iz := int64(int32(z))
	part1 := 116129781 * iz
	part2 := int64(0x2FC20F00000001*ux) >> 32
	v1 := part1 ^ part2
	calc := v1 * (42317861*v1 + 11)
	temp := uint64(calc) >> 16
	signExtended := int64(int32(temp))
	return uint64(signExtended) ^ 0x6A09E667F3BCC909
}

// randomToFloat32 convert random long to float in [0, 1).
func (Bamboo) randomToFloat32(random uint64) float32 {
	return float32(random>>40) * (1.0 / 16777216.0)
}

// calculateOffsetValue calculate offset value with quantization to discrete steps.
func (Bamboo) calculateOffsetValue(mn, mx float32, steps int, random float32) float32 {
	if mn >= mx {
		return mn
	}
	if steps == 1 {
		return (mn + mx) * 0.5
	} else if steps > 1 {
		rng := mx - mn
		stepSize := rng / float32(steps-1)
		val := float32(steps) * random
		index := float32(math.Floor(float64(val)))
		return mn + index*stepSize
	}
	return mn + (mx-mn)*random
}

// randomlyModifyPosition ...
func (b Bamboo) randomlyModifyPosition(pos cube.Pos) mgl64.Vec3 {
	seed := b.positionHash(pos.X(), pos.Z())
	s0 := mcrandom.MixStafford13(seed)
	s1 := mcrandom.MixStafford13(seed + 0x9e3779b97f4a7c15)
	prng := mcrandom.NewXoroshiro128PlusPlus(s0, s1)
	offsetX := b.calculateOffsetValue(-0.25, 0.25, 16, b.randomToFloat32(prng.Next()))
	prng.Next()
	offsetZ := b.calculateOffsetValue(-0.25, 0.25, 16, b.randomToFloat32(prng.Next()))
	return mgl64.Vec3{float64(offsetX), 0, float64(offsetZ)}
}

// FaceSolid ...
func (b Bamboo) FaceSolid(pos cube.Pos, face cube.Face, s world.BlockSource) bool {
	return false
}
