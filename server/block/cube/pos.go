package cube

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Pos holds the position of a block. The position is represented of an array with an x, y and z value,
// where the y value is positive.
type Pos [3]int

// String converts the Pos to a string in the format (1,2,3) and returns it.
func (p Pos) String() string {
	return fmt.Sprintf("(%v,%v,%v)", p[0], p[1], p[2])
}

// X returns the X coordinate of the block position.
func (p Pos) X() int {
	return p[0]
}

// Y returns the Y coordinate of the block position.
func (p Pos) Y() int {
	return p[1]
}

// Z returns the Z coordinate of the block position.
func (p Pos) Z() int {
	return p[2]
}

// OutOfBounds checks if the Y value is either bigger than r[1] or smaller than r[0].
func (p Pos) OutOfBounds(r Range) bool {
	y := p[1]
	return y > r[1] || y < r[0]
}

// Add adds two block positions together and returns a new one with the combined values.
func (p Pos) Add(pos Pos) Pos {
	return Pos{p[0] + pos[0], p[1] + pos[1], p[2] + pos[2]}
}

// Sub subtracts pos from p and returns a new one with the subtracted values.
func (p Pos) Sub(pos Pos) Pos {
	return Pos{p[0] - pos[0], p[1] - pos[1], p[2] - pos[2]}
}

// Vec3 returns a vec3 holding the same coordinates as the block position.
func (p Pos) Vec3() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]), float64(p[1]), float64(p[2])}
}

// Vec3Middle returns a Vec3 holding the coordinates of the block position with 0.5 added on both horizontal
// axes.
func (p Pos) Vec3Middle() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]) + 0.5, float64(p[1]), float64(p[2]) + 0.5}
}

// Vec3Centre returns a Vec3 holding the coordinates of the block position with 0.5 added on all axes.
func (p Pos) Vec3Centre() mgl64.Vec3 {
	return mgl64.Vec3{float64(p[0]) + 0.5, float64(p[1]) + 0.5, float64(p[2]) + 0.5}
}

// Side returns the position on the side of this block position, at a specific face.
func (p Pos) Side(face Face) Pos {
	switch face {
	case FaceUp:
		p[1]++
	case FaceDown:
		p[1]--
	case FaceNorth:
		p[2]--
	case FaceSouth:
		p[2]++
	case FaceWest:
		p[0]--
	case FaceEast:
		p[0]++
	}
	return p
}

// Face returns the face that the other Pos was on compared to the current Pos. The other Pos
// is assumed to be a direct neighbour of the current Pos.
func (p Pos) Face(other Pos) Face {
	switch other {
	case p.Add(Pos{0, 1}):
		return FaceUp
	case p.Add(Pos{0, -1}):
		return FaceDown
	case p.Add(Pos{0, 0, -1}):
		return FaceNorth
	case p.Add(Pos{0, 0, 1}):
		return FaceSouth
	case p.Add(Pos{-1, 0, 0}):
		return FaceWest
	case p.Add(Pos{1, 0, 0}):
		return FaceEast
	}
	return FaceUp
}

// Neighbours calls the function passed for each of the block position's neighbours. If the Y value is out of
// bounds, the function will not be called for that position.
func (p Pos) Neighbours(f func(neighbour Pos), r Range) {
	if p.OutOfBounds(r) {
		return
	}
	p[0]++
	f(p)
	p[0] -= 2
	f(p)
	p[0]++
	p[1]++
	if p[1] <= r[1] {
		f(p)
	}
	p[1] -= 2
	if p[1] >= r[0] {
		f(p)
	}
	p[1]++
	p[2]++
	f(p)
	p[2] -= 2
	f(p)
}

// PosFromVec3 returns a block position by a Vec3, rounding the values down adequately.
func PosFromVec3(vec3 mgl64.Vec3) Pos {
	return Pos{int(math.Floor(vec3[0])), int(math.Floor(vec3[1])), int(math.Floor(vec3[2]))}
}
