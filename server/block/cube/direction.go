package cube

// Direction represents a direction towards one of the horizontal axes of the world.
type Direction int

const (
	// North represents the north direction, towards the negative Z.
	North Direction = iota
	// South represents the south direction, towards the positive Z.
	South
	// West represents the west direction, towards the negative X.
	West
	// East represents the east direction, towards the positive X.
	East
)

// Face converts the direction to a Face and returns it.
func (d Direction) Face() Face {
	return Face(d + 2)
}

// Opposite returns Direction opposite to the current one.
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

// RotateRight rotates the direction 90 degrees to the right horizontally and returns the new direction.
func (d Direction) RotateRight() Direction {
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

// RotateLeft rotates the direction 90 degrees to the left horizontally and returns the new direction.
func (d Direction) RotateLeft() Direction {
	switch d {
	case North:
		return West
	case East:
		return North
	case South:
		return East
	case West:
		return South
	}
	panic("invalid direction")
}

// String returns the Direction as a string.
func (d Direction) String() string {
	switch d {
	case North:
		return "north"
	case East:
		return "east"
	case South:
		return "south"
	case West:
		return "west"
	}
	panic("invalid direction")
}

var directions = [...]Direction{North, East, South, West}

// Directions returns a list of all directions, going from North to West.
func Directions() []Direction {
	return directions[:]
}
