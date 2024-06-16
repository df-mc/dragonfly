package block

// DoubleTallGrassType represents a type of double tall grass, which can be placed on top of grass blocks.
type DoubleTallGrassType struct {
	doubleTallGrass
}

// NormalDoubleTallGrass returns the normal variant of double tall grass.
func NormalDoubleTallGrass() DoubleTallGrassType {
	return DoubleTallGrassType{0}
}

// FernDoubleTallGrass returns the fern variant of double tall grass.
func FernDoubleTallGrass() DoubleTallGrassType {
	return DoubleTallGrassType{1}
}

// DoubleTallGrassTypes returns all variants of double tall grass.
func DoubleTallGrassTypes() []DoubleTallGrassType {
	return []DoubleTallGrassType{NormalDoubleTallGrass(), FernDoubleTallGrass()}
}

type doubleTallGrass uint8

// Uint8 ...
func (t doubleTallGrass) Uint8() uint8 {
	return uint8(t)
}

// Name ...
func (t doubleTallGrass) Name() string {
	switch t {
	case 0:
		return "Tall Grass"
	case 1:
		return "Large Fern"
	}
	panic("unknown double tall grass type")
}

// String ...
func (t doubleTallGrass) String() string {
	switch t {
	case 0:
		return "tall_grass"
	case 1:
		return "large_fern"
	}
	panic("unknown double tall grass type")
}
