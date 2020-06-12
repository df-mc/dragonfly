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
