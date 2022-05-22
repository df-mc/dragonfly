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
	}
	panic("should never happen")
}
