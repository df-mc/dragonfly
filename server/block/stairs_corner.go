package block

// StairsCorner represents the corner that stairs form with the stairs around
// them. It is calculated by the server whenever the stairs or the stairs around
// them change.
type StairsCorner struct {
	stairsCorner
}

// NoStairsCorner returns the corner of stairs that do not form a corner with any of the stairs around them.
func NoStairsCorner() StairsCorner {
	return StairsCorner{0}
}

// InnerLeftStairsCorner returns the corner of stairs that turn inwards on their left side.
func InnerLeftStairsCorner() StairsCorner {
	return StairsCorner{1}
}

// InnerRightStairsCorner returns the corner of stairs that turn inwards on their right side.
func InnerRightStairsCorner() StairsCorner {
	return StairsCorner{2}
}

// OuterLeftStairsCorner returns the corner of stairs that turn outwards on their left side.
func OuterLeftStairsCorner() StairsCorner {
	return StairsCorner{3}
}

// OuterRightStairsCorner returns the corner of stairs that turn outwards on their right side.
func OuterRightStairsCorner() StairsCorner {
	return StairsCorner{4}
}

// StairsCorners returns a list of all stairs corners.
func StairsCorners() []StairsCorner {
	return []StairsCorner{NoStairsCorner(), InnerLeftStairsCorner(), InnerRightStairsCorner(), OuterLeftStairsCorner(), OuterRightStairsCorner()}
}

type stairsCorner uint8

// Uint8 returns the stairs corner as a uint8.
func (s stairsCorner) Uint8() uint8 {
	return uint8(s)
}

// Inner returns true if the corner turns inwards, filling up the block rather than a quarter of it.
func (s stairsCorner) Inner() bool {
	return s == 1 || s == 2
}

// Left returns true if the corner is on the left side of the stairs.
func (s stairsCorner) Left() bool {
	return s == 1 || s == 3
}

// String ...
func (s stairsCorner) String() string {
	switch s {
	case 0:
		return "none"
	case 1:
		return "inner_left"
	case 2:
		return "inner_right"
	case 3:
		return "outer_left"
	case 4:
		return "outer_right"
	}
	panic("should never happen")
}
