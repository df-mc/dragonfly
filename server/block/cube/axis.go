package cube

import "github.com/go-gl/mathgl/mgl64"

// Axis represents the axis that a block, such as a log, may be directed in.
type Axis int

const (
	// Y represents the vertical Y axis.
	Y Axis = iota
	// Z represents the horizontal Z axis.
	Z
	// X represents the horizontal X axis.
	X
)

// String converts an Axis into either x, y or z, depending on which axis it is.
func (a Axis) String() string {
	switch a {
	case X:
		return "x"
	case Y:
		return "y"
	default:
		return "z"
	}
}

// RotateLeft rotates an Axis from X to Z or from Z to X.
func (a Axis) RotateLeft() Axis {
	switch a {
	case X:
		return Z
	case Z:
		return X
	default:
		return 0
	}
}

// RotateRight rotates an Axis from X to Z or from Z to X.
func (a Axis) RotateRight() Axis {
	// No difference in rotating left or right for an Axis.
	return a.RotateLeft()
}

// Vec3 returns a unit Vec3 of either (1, 0, 0), (0, 1, 0) or (0, 0, 1),
// depending on the Axis.
func (a Axis) Vec3() mgl64.Vec3 {
	switch a {
	case X:
		return mgl64.Vec3{1, 0, 0}
	case Y:
		return mgl64.Vec3{0, 1, 0}
	default:
		return mgl64.Vec3{0, 0, 1}
	}
}

// Axes return all possible axes. (x, y, z)
func Axes() []Axis {
	return []Axis{X, Y, Z}
}
