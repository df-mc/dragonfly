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
	seed := offsetSeed(int32(pos.X()), int32(pos.Z()))
	s0 := mcrandom.MixStafford13(seed)
	s1 := mcrandom.MixStafford13(seed + 0x9E3779B97F4A7C15)
	prng := mcrandom.NewXoroshiro128PlusPlus(s0, s1)
	x := offsetValue(mn, mx, steps, randomFloat32(prng.Next()))
	prng.Next() // The Y offset is always zero, but its draw is still consumed.
	z := offsetValue(mn, mx, steps, randomFloat32(prng.Next()))
	return mgl64.Vec3{float64(x), 0, float64(z)}
}

// offsetSeed returns the seed vanilla uses for random model offsets. Both coordinate
// products use signed 64-bit arithmetic. The signed 32-bit extraction after the
// nonlinear mix is intentional.
func offsetSeed(x, z int32) uint64 {
	v := int64(z)*116129781 ^ int64(x)*0x2fc20f
	v = int64(int32(uint64(v*(v*42317861+11)) >> 16))
	return uint64(v) ^ 0x6A09E667F3BCC909
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
