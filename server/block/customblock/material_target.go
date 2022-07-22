package customblock

// Target represents a material target for a custom block. These are limited to either all targets, or the top,
// bottom, and sides.
type Target struct {
	materialTarget
}

// MaterialTargetAll represents the material target for all targets.
func MaterialTargetAll() Target {
	return Target{0}
}

// MaterialTargetUp represents the material target for the top of the block.
func MaterialTargetUp() Target {
	return Target{1}
}

// MaterialTargetDown represents the material target for the bottom of the block.
func MaterialTargetDown() Target {
	return Target{2}
}

// MaterialTargetSides represents the material target for the sides of the block.
func MaterialTargetSides() Target {
	return Target{3}
}

// MaterialTargetNorth represents the material target for the north face of the block.
func MaterialTargetNorth() Target {
	return Target{4}
}

// MaterialTargetEast represents the material target for the east face of the block.
func MaterialTargetEast() Target {
	return Target{5}
}

// MaterialTargetSouth represents the material target for the south face of the block.
func MaterialTargetSouth() Target {
	return Target{6}
}

// MaterialTargetWest represents the material target for the west face of the block.
func MaterialTargetWest() Target {
	return Target{7}
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
