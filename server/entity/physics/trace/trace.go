package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// TraverseBlocks performs a ray trace between the start and end coordinates.
// A function 'f' is passed which is called for each voxel, if f returns false, the function will return.
// TraverseBlocks panics if the start and end positions are the same.
func TraverseBlocks(start, end mgl64.Vec3, f func(pos cube.Pos) (con bool)) {
	dir := end.Sub(start).Normalize()
	if dir.LenSqr() <= 0.0 {
		panic("start and end points are the same, giving a zero direction vector")
	}

	b := cube.PosFromVec3(start)

	step := signVec3(dir)
	stepX, stepY, stepZ := cube.Pos{int(math.Floor(step[0]))}, cube.Pos{0, int(math.Floor(step[1]))}, cube.Pos{0, 0, int(math.Floor(step[2]))}
	max := boundaryVec3(start, dir)

	delta := divideVec3(step, dir)

	r := world.Distance(start, end)
	for {
		if !f(b) {
			return
		}

		if max[0] < max[1] && max[0] < max[2] {
			if max[0] > r {
				return
			}
			b = b.Add(stepX)
			max[0] += delta[0]
		} else if max[1] < max[2] {
			if max[1] > r {
				return
			}
			b = b.Add(stepY)
			max[1] += delta[1]
		} else {
			if max[2] > r {
				return
			}
			b = b.Add(stepZ)
			max[2] += delta[2]
		}
	}
}

// divideVec3 ...
func divideVec3(v1, v2 mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{divide(v1[0], v2[0]), divide(v1[1], v2[1]), divide(v1[2], v2[2])}
}

// divide ...
func divide(f1, f2 float64) float64 {
	if f2 == 0.0 {
		return 0.0
	}
	return f1 / f2
}

// boundaryVec3 ...
func boundaryVec3(v1, v2 mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{boundary(v1[0], v2[0]), boundary(v1[1], v2[1]), boundary(v1[2], v2[2])}
}

// boundary ...
func boundary(s, d float64) float64 {
	if d == 0.0 {
		return math.Inf(1)
	}

	if d < 0.0 {
		s, d = -s, -d
		if math.Floor(s) == s {
			return 0.0
		}
	}

	return (1 - (s - math.Floor(s))) / d
}

// signVec3 ...
func signVec3(v1 mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{sign(v1[0]), sign(v1[1]), sign(v1[2])}
}

// sign ...
func sign(f float64) float64 {
	switch {
	case f > 0.0:
		return 1.0
	case f < 0.0:
		return -1.0
	}
	return 0.0
}
