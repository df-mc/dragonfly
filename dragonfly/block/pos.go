package block

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Position holds the position of a block. The position is represented of an array with an x, y and z value,
// where the y value is positive.
type Position [3]int

// X returns the X coordinate of the block position.
func (p Position) X() int {
	return p[0]
}

// Y returns the Y coordinate of the block position.
func (p Position) Y() int {
	return p[1]
}

// Z returns the Z coordinate of the block position.
func (p Position) Z() int {
	return p[2]
}

// Vec3 returns a vec3 holding the same coordinates as the block position.
func (p Position) Vec3() mgl32.Vec3 {
	return mgl32.Vec3{float32(p[0]), float32(p[1]), float32(p[2])}
}

// Side returns the position on the side of this block position, at a specific face.
func (p Position) Side(face Face) Position {
	switch face {
	case Up:
		p[1]++
	case Down:
		p[1]--
	case North:
		p[2]--
	case South:
		p[2]++
	case West:
		p[0]--
	case East:
		p[0]++
	}
	return p
}
