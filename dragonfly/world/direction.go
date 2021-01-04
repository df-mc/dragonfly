package world

// Direction represents a direction towards one of the horizontal axes of the world.
type Direction int

const (
	// North represents the north direction.
	North Direction = iota
	// South represents the south direction.
	South
	// West represents the west direction.
	West
	// East represents the east direction.
	East
)

// Opposite returns the opposite direction.
func (d Direction) Opposite() Direction {
	switch d {
	case North:
		return South
	case South:
		return North
	case West:
		return East
	case East:
		return West
	}
	panic("invalid direction")
}

// Face converts the direction to a block face.
func (d Direction) Face() Face {
	return Face(d + 2)
}

// Rotate90 rotates the direction 90 degrees horizontally and returns the new direction.
func (d Direction) Rotate90() Direction {
	switch d {
	case North:
		return East
	case East:
		return South
	case South:
		return West
	case West:
		return North
	}
	panic("invalid direction")
}

func AllDirections() (d []Direction) {
	d = make([]Direction, 4)

	d[0] = South
	d[1] = West
	d[2] = North
	d[3] = East

	return
}
