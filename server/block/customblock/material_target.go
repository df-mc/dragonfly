package customblock

// MaterialTarget represents a material target for a custom block. These are limited to either all targets, or the top,
// bottom, and sides.
type MaterialTarget struct {
	materialTarget
}

// MaterialTargetAll represents the material target for all targets.
func MaterialTargetAll() MaterialTarget {
	return MaterialTarget{0}
}

// MaterialTargetUp represents the material target for the top of the block.
func MaterialTargetUp() MaterialTarget {
	return MaterialTarget{1}
}

// MaterialTargetDown represents the material target for the bottom of the block.
func MaterialTargetDown() MaterialTarget {
	return MaterialTarget{2}
}

// MaterialTargetSides represents the material target for the sides of the block.
func MaterialTargetSides() MaterialTarget {
	return MaterialTarget{3}
}

// MaterialTargetNorth represents the material target for the north face of the block.
func MaterialTargetNorth() MaterialTarget {
	return MaterialTarget{4}
}

// MaterialTargetEast represents the material target for the east face of the block.
func MaterialTargetEast() MaterialTarget {
	return MaterialTarget{5}
}

// MaterialTargetSouth represents the material target for the south face of the block.
func MaterialTargetSouth() MaterialTarget {
	return MaterialTarget{6}
}

// MaterialTargetWest represents the material target for the west face of the block.
func MaterialTargetWest() MaterialTarget {
	return MaterialTarget{7}
}

type materialTarget uint8

// Name returns the name of the material target.
func (m materialTarget) Name() string {
	switch m {
	case 0:
		return "all"
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "sides"
	case 4:
		return "north"
	case 5:
		return "east"
	case 6:
		return "south"
	case 7:
		return "west"
	}
	panic("should never happen")
}

// String returns the string representation of the material target.
func (m materialTarget) String() string {
	switch m {
	case 0:
		return "*"
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "sides"
	case 4:
		return "north"
	case 5:
		return "east"
	case 6:
		return "south"
	case 7:
		return "west"
	}
	panic("should never happen")
}
