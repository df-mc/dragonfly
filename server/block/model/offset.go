package model

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/mcrandom"
	"github.com/go-gl/mathgl/mgl64"
)

// randomOffset returns the deterministic horizontal offset vanilla applies to
// the models of certain blocks, such as bamboo, at the position passed.
func randomOffset(pos cube.Pos, mn, mx float32, steps int) mgl64.Vec3 {
	l := uint64(getSeed(int32(pos.X()), 0, int32(pos.Z()))) ^ 0x6A09E667F3BCC909
	s0 := mcrandom.MixStafford13(l)
	s1 := mcrandom.MixStafford13(l + 0x9E3779B97F4A7C15)
	prng := mcrandom.NewXoroshiro128PlusPlus(s0, s1)
	x := offsetValue(mn, mx, steps, randomFloat32(prng.Next()))
	prng.Next() // Y offset, unused.
	z := offsetValue(mn, mx, steps, randomFloat32(prng.Next()))
	return mgl64.Vec3{float64(x), 0, float64(z)}
}

// getSeed returns the position hash vanilla uses for random model offsets.
func getSeed(x, y, z int32) uint32 {
	v := uint32(y) ^ 3129871*uint32(x) ^ 116129781*uint32(z)
	return v * (42317861*v + 11)
}

// randomFloat32 converts a random uint64 to a float in [0, 1).
func randomFloat32(random uint64) float32 {
	return float32(random>>40) * (1.0 / 16777216.0)
}

// offsetValue quantizes a random float to one of steps values in [mn, mx].
func offsetValue(mn, mx float32, steps int, random float32) float32 {
	if mn >= mx {
		return mn
	}
	if steps == 1 {
		return (mn + mx) * 0.5
	}
	if steps > 1 {
		index := float32(math.Floor(float64(float32(steps) * random)))
		return mn + index*(mx-mn)/float32(steps-1)
	}
	return mn + (mx-mn)*random
}
