package world

import "fmt"

// Face represents the face of a block or entity.
type Face int

// FromString returns a Face by a string.
func (f Face) FromString(s string) (interface{}, error) {
	switch s {
	case "down":
		return FaceDown, nil
	case "up":
		return FaceUp, nil
	case "north":
		return FaceNorth, nil
	case "south":
		return FaceSouth, nil
	case "west":
		return FaceWest, nil
	case "east":
		return FaceEast, nil
	}
	return nil, fmt.Errorf("unexpected facing '%v', expecting one of 'down', 'up', 'north', 'south', 'west' or 'east'", s)
}

const (
	// FaceDown represents the bottom face of a block.
	FaceDown Face = iota
	// FaceUp represents the top face of a block.
	FaceUp
	// FaceNorth represents the north face of a block.
	FaceNorth
	// FaceSouth represents the south face of a block.
	FaceSouth
	// FaceWest represents the west face of the block.
	FaceWest
	// FaceEast represents the east face of the block.
	FaceEast
)

// Opposite returns the opposite face. FaceDown will return up, north will return south and west will return east,
// and vice versa.
func (f Face) Opposite() Face {
	switch f {
	default:
		return FaceUp
	case FaceUp:
		return FaceDown
	case FaceNorth:
		return FaceSouth
	case FaceSouth:
		return FaceNorth
	case FaceWest:
		return FaceEast
	case FaceEast:
		return FaceWest
	}
}

// Axis returns the axis the face is facing. FaceEast and west correspond to the x axis, north and south to the z
// axis and up and down to the y axis.
func (f Face) Axis() Axis {
	switch f {
	default:
		return Y
	case FaceEast, FaceWest:
		return X
	case FaceNorth, FaceSouth:
		return Z
	}
}
