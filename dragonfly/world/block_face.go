package world

// Face represents the face of a block or entity.
type Face int

const (
	// Down represents the bottom face of a block.
	Down Face = iota
	// Up represents the top face of a block.
	Up
	// North represents the north face of a block.
	North
	// South represents the south face of a block.
	South
	// West represents the west face of the block.
	West
	// East represents the east face of the block.
	East
)

// Opposite returns the opposite face. Down will return up, north will return south and west will return east,
// and vice versa.
func (f Face) Opposite() Face {
	switch f {
	default:
		return Up
	case Up:
		return Down
	case North:
		return South
	case South:
		return North
	case West:
		return East
	case East:
		return West
	}
}

// Axis returns the axis the face is facing. East and west correspond to the x axis, north and south to the z
// axis and up and down to the y axis.
func (f Face) Axis() Axis {
	switch f {
	default:
		return Y
	case East, West:
		return X
	case North, South:
		return Z
	}
}
