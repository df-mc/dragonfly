package grass

// TallGrass represents the tall grass block.
type TallGrass struct {
	grass
}

type grass uint8

// Default returns the default type of tall grass.
func Default() TallGrass {
	return TallGrass{grass(0)}
}

// Tall returns the tall type of tall grass.
func Tall() TallGrass {
	return TallGrass{grass(1)}
}

// Fern returns the fern type of tall grass.
func Fern() TallGrass {
	return TallGrass{grass(2)}
}

// Snow returns the snow type of tall grass.
func Snow() TallGrass {
	return TallGrass{grass(3)}
}

// All returns all types of tall grass.
func All() []TallGrass {
	return []TallGrass{Default(), Tall(), Fern(), Snow()}
}

// Name returns the name of the tall grass
func (g TallGrass) Name() string {
	switch g.grass {
	case 0:
		return "default"
	case 1:
		return "tall"
	case 2:
		return "fern"
	case 3:
		return "snow"
	}
	panic("unknown tall grass type")
}
