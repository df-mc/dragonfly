package block

// GrassType represents a grass plant, which can be placed on top of grass blocks.
type GrassType struct {
	grass
}

// NormalGrass returns the grass variant of grass.
func NormalGrass() GrassType {
	return GrassType{0}
}

// Fern returns the fern variant of grass.
func Fern() GrassType {
	return GrassType{1}
}

// GrassTypes returns all variants of grass.
func GrassTypes() []GrassType {
	return []GrassType{NormalGrass(), Fern()}
}

type grass uint8

// Uint8 converts the grass to an integer that uniquely identifies its type.
func (g grass) Uint8() uint8 {
	return uint8(g)
}

// Name returns the grass's display name.
func (g grass) Name() string {
	switch g {
	case 0:
		return "Grass"
	case 1:
		return "Fern"
	}
	panic("unknown grass type")
}

// String ...
func (g grass) String() string {
	switch g {
	case 0:
		return "grass"
	case 1:
		return "fern"
	}
	panic("unknown grass type")
}
