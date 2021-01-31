package grass

import (
	"fmt"
)

// Grass represents a grass plant, which can be placed on top of grass blocks.
type Grass struct {
	grass
}

// SmallGrass returns the small grass variant of grass.
func SmallGrass() Grass {
	return Grass{0}
}

// Fern returns the fern variant of grass.
func Fern() Grass {
	return Grass{1}
}

// TallGrass returns the tall grass variant of grass.
func TallGrass() Grass {
	return Grass{2}
}

// LargeFern returns the large fern variant of grass.
func LargeFern() Grass {
	return Grass{3}
}

// NetherSprouts returns the nether sprouts variant of grass.
func NetherSprouts() Grass {
	return Grass{4}
}

// All returns all variants of grass.
func All() []Grass {
	return []Grass{SmallGrass(), Fern(), TallGrass(), LargeFern(), NetherSprouts()}
}

type grass uint8

// Uint8 converts the grass to an integer that uniquely identifies it's type.
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
	case 2:
		return "Tall Grass"
	case 3:
		return "Large Fern"
	case 4:
		return "Nether Sprouts"
	}
	panic("unknown grass type")
}

// FromString ...
func (g grass) FromString(s string) (interface{}, error) {
	switch s {
	case "grass":
		return SmallGrass(), nil
	case "fern":
		return Fern(), nil
	case "tall grass":
		return TallGrass(), nil
	case "large fern":
		return LargeFern(), nil
	case "nether sprouts":
		return NetherSprouts(), nil
	}
	return nil, fmt.Errorf("unexpected grass type '%v', expecting one of 'grass', 'fern', or 'tall grass'", s)
}

// String ...
func (g grass) String() string {
	switch g {
	case 0:
		return "grass"
	case 1:
		return "fern"
	case 2:
		return "tall grass"
	case 3:
		return "large fern"
	case 4:
		return "nether sprouts"
	}
	panic("unknown grass type")
}
