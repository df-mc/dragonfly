package grass

import (
	"fmt"
)

type Grass struct {
	grass
}

func SmallGrass() Grass {
	return Grass{0}
}

func Fern() Grass {
	return Grass{1}
}

func TallGrass() Grass {
	return Grass{2}
}

func LargeFern() Grass {
	return Grass{3}
}

func NetherSprouts() Grass {
	return Grass{4}
}

func All() []Grass {
	return []Grass{SmallGrass(), Fern(), TallGrass(), LargeFern(), NetherSprouts()}
}

type grass uint8

func (g grass) Uint8() uint8 {
	return uint8(g)
}

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
