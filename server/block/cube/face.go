package cube

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

// Face represents the face of a block or entity.
type Face int

// Direction converts the Face to a Direction and returns it, assuming the Face
// is horizontal and not FaceUp or FaceDown.
func (f Face) Direction() Direction {
	return Direction(f - 2)
}

// Positive returns whether the face is the positive of its axis. For example, FaceEast is positive, FaceWest is not.
func (f Face) Positive() bool {
	switch f {
	case FaceUp, FaceSouth, FaceEast:
		return true
	}
	return false
}

// Opposite returns the opposite Face. FaceDown will return FaceUp, FaceNorth
// will return FaceSouth, FaceWest will return FaceEast, and vice versa.
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

// Axis returns the Axis the face is facing. FaceEast and FaceWest correspond
// to the X Axis, FaceNorth and FaceSouth to the Z Axis and FaceUp and FaceDown
// to the Y Axis.
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

// RotateRight rotates the face 90 degrees to the right horizontally
// (clockwise) and returns the new Face.
func (f Face) RotateRight() Face {
	switch f {
	case FaceNorth:
		return FaceEast
	case FaceEast:
		return FaceSouth
	case FaceSouth:
		return FaceWest
	case FaceWest:
		return FaceNorth
	}
	return f
}

// RotateLeft rotates the face 90 degrees to the left horizontally
// (counter-clockwise) and returns the new Face.
func (f Face) RotateLeft() Face {
	switch f {
	case FaceNorth:
		return FaceWest
	case FaceEast:
		return FaceNorth
	case FaceSouth:
		return FaceEast
	case FaceWest:
		return FaceSouth
	}
	return f
}

// String returns the Face as a string.
func (f Face) String() string {
	switch f {
	case FaceUp:
		return "up"
	case FaceDown:
		return "down"
	case FaceNorth:
		return "north"
	case FaceSouth:
		return "south"
	case FaceWest:
		return "west"
	case FaceEast:
		return "east"
	}
	panic("invalid face")
}

// Faces returns a list of all faces, starting with down, then up, then north
// to west.
func Faces() []Face {
	return faces[:]
}

// HorizontalFaces returns a list of all horizontal faces, from north to west.
func HorizontalFaces() []Face {
	return hFaces[:]
}

var hFaces = [...]Face{FaceNorth, FaceEast, FaceSouth, FaceWest}

var faces = [...]Face{FaceDown, FaceUp, FaceNorth, FaceEast, FaceSouth, FaceWest}
