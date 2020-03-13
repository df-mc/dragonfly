package world

// Axis represents the axis that a block may be directed in. Most blocks do not have an axis, but blocks such
// as logs or pillars do.
type Axis int

const (
	// Y represents the vertical Y axis.
	Y Axis = iota
	// Z represents the horizontal Z axis.
	Z
	// X represents the horizontal X axis.
	X
)

// EncodeBlock converts an Axis into either x, y or z, depending on which axis it is.
func (a Axis) String() string {
	if a == X {
		return "x"
	} else if a == Y {
		return "y"
	}
	return "z"
}
