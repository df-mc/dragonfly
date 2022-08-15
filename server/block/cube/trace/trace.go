package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// TraverseBlocks performs a ray trace between the start and end coordinates.
// A function 'f' is passed which is called for each voxel, if f returns false, the function will return.
// TraverseBlocks panics if the start and end positions are the same.
func TraverseBlocks(start, end mgl64.Vec3, f func(pos cube.Pos) (con bool)) {
	dir := end.Sub(start)
	if mgl64.FloatEqual(dir.LenSqr(), 0) {
		panic("start and end points are the same, giving a zero direction vector")
	}
	dir = dir.Normalize()

	b := cube.PosFromVec3(start)

	step := signVec3(dir)
	stepX, stepY, stepZ := int(step[0]), int(step[1]), int(step[2])
	max := boundaryVec3(start, dir)

	delta := safeDivideVec3(step, dir)

	r := start.Sub(end).Len()
	for {
		if !f(b) {
			return
		}

		if max[0] < max[1] && max[0] < max[2] {
			if max[0] > r {
				return
			}
			b[0] += stepX
			max[0] += delta[0]
		} else if max[1] < max[2] {
			if max[1] > r {
				return
			}
			b[1] += stepY
			max[1] += delta[1]
		} else {
			if max[2] > r {
				return
			}
			b[2] += stepZ
			max[2] += delta[2]
		}
	}
}

// safeDivideVec3 ...
func safeDivideVec3(dividend, divisor mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{
		safeDivide(dividend[0], divisor[0]),
		safeDivide(dividend[1], divisor[1]),
		safeDivide(dividend[2], divisor[2]),
	}
}

// safeDivide divides the dividend by the divisor, but if the divisor is 0, it returns 0.
func safeDivide(dividend, divisor float64) float64 {
	if divisor == 0.0 {
		return 0.0
	}
	return dividend / divisor
}

// boundaryVec3 ...
func boundaryVec3(v1, v2 mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{boundary(v1[0], v2[0]), boundary(v1[1], v2[1]), boundary(v1[2], v2[2])}
}

// boundary returns the distance that must be travelled on an axis from the start point with the direction vector
// component to cross a block boundary.
func boundary(start, dir float64) float64 {
	if dir == 0.0 {
		return math.Inf(1)
	}

	if dir < 0.0 {
		start, dir = -start, -dir
		if math.Floor(start) == start {
			return 0.0
		}
	}

	return (1 - (start - math.Floor(start))) / dir
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
