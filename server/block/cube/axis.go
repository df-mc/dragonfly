package cube

import (
	"fmt"
)

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

// String converts an Axis into either x, y or z, depending on which axis it is.
func (a Axis) String() string {
	if a == X {
		return "x"
	} else if a == Y {
		return "y"
	}
	return "z"
}

// FromString returns an axis by a string.
func (a Axis) FromString(s string) (interface{}, error) {
	switch s {
	case "x":
		return X, nil
	case "y":
		return Y, nil
	case "z":
		return Z, nil
	}
	return nil, fmt.Errorf("unexpected axis '%v', expecting one of 'x', 'y' or 'z'", s)
}

// Axes return all possible axes. (x, y, z)
func Axes() []Axis {
	return []Axis{X, Y, Z}
}
