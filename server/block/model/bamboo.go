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
	// The stalk's box extends from the block's centre towards positive X and Z.
	size := 0.5 + 2.0/16.0
	if b.Thick {
		size = 0.5 + 3.0/16.0
	}
	return []cube.BBox{cube.Box(0.5, 0, 0.5, size, 1, size).Translate(b.randomlyModifyPosition(pos))}
}

// getSeed returns the position hash vanilla uses for the stalk's random offset.
func getSeed(x, y, z int32) uint32 {
	v := uint32(y) ^ 3129871*uint32(x) ^ 116129781*uint32(z)
	return v * (42317861*v + 11)
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

// randomlyModifyPosition returns the horizontal offset of the stalk at pos.
// TODO: Verify the xoroshiro seeding and offset bounds against vanilla.
func (b Bamboo) randomlyModifyPosition(pos cube.Pos) mgl64.Vec3 {
	l := uint64(getSeed(int32(pos.X()), 0, int32(pos.Z()))) ^ 0x6A09E667F3BCC909
	s0 := mcrandom.MixStafford13(l)
	s1 := mcrandom.MixStafford13(l + 0x9E3779B97F4A7C15)
	prng := mcrandom.NewXoroshiro128PlusPlus(s0, s1)
	offsetX := b.calculateOffsetValue(-0.25, 0.25, 16, b.randomToFloat32(prng.Next()))
	prng.Next() // Y offset, unused.
	offsetZ := b.calculateOffsetValue(-0.25, 0.25, 16, b.randomToFloat32(prng.Next()))
	return mgl64.Vec3{float64(offsetX), 0, float64(offsetZ)}
}

// FaceSolid ...
func (b Bamboo) FaceSolid(pos cube.Pos, face cube.Face, s world.BlockSource) bool {
	return false
}
