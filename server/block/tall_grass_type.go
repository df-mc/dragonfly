package block

// TallGrassType represents a tall grass plant, which can be placed on top of grass blocks.
type TallGrassType struct {
	tallGrass
}

// LegacyTallGrass returns the legacy tall grass variant of tall grass.
func LegacyTallGrass() TallGrassType {
	return TallGrassType{0}
}

// NormalTallGrass returns the tall grass variant of tall grass.
func NormalTallGrass() TallGrassType {
	return TallGrassType{1}
}

// FernTallGrass returns the fern variant of tall grass.
func FernTallGrass() TallGrassType {
	return TallGrassType{2}
}

// TallGrassTypes returns all variants of tall grass.
func TallGrassTypes() []TallGrassType {
	return []TallGrassType{LegacyTallGrass(), NormalTallGrass(), FernTallGrass()}
}

type tallGrass uint8

// Uint8 ...
func (g tallGrass) Uint8() uint8 {
	return uint8(g)
}

// Double returns the double tall grass variant of the tall grass.
func (g tallGrass) Double() DoubleTallGrassType {
	switch g {
	case 0, 1:
		return NormalDoubleTallGrass()
	case 2:
		return FernDoubleTallGrass()
	}
	panic("unknown tall grass type")
}

// Name ...
func (g tallGrass) Name() string {
	switch g {
	case 0, 1:
		return "Grass"
	case 2:
		return "Fern"
	}
	panic("unknown tall grass type")
}

// String ...
func (g tallGrass) String() string {
	switch g {
	case 0:
		return "default"
	case 1:
		return "tall"
	case 2:
		return "fern"
	}
	panic("unknown tall grass type")
}
