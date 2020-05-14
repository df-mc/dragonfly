package material

import "fmt"

// Wood represents a type of wood of a block. Some blocks, such as log blocks, bark blocks, wooden planks and
// others carry one of these types.
type Wood struct {
	wood
}

// OakWood returns oak wood material.
func OakWood() Wood {
	return Wood{wood(0)}
}

// SpruceWood returns spruce wood material.
func SpruceWood() Wood {
	return Wood{wood(1)}
}

// BirchWood returns birch wood material.
func BirchWood() Wood {
	return Wood{wood(2)}
}

// JungleWood returns jungle wood material.
func JungleWood() Wood {
	return Wood{wood(3)}
}

// AcaciaWood returns acacia wood material.
func AcaciaWood() Wood {
	return Wood{wood(4)}
}

// DarkOakWood returns dark oak wood material.
func DarkOakWood() Wood {
	return Wood{wood(5)}
}

type wood uint8

// Name ...
func (w wood) Name() string {
	switch w {
	case 0:
		return "Oak"
	case 1:
		return "Spruce"
	case 2:
		return "Birch"
	case 3:
		return "Jungle"
	case 4:
		return "Acacia"
	case 5:
		return "Dark Oak"
	}
	panic("unknown wood type")
}

// FromString ...
func (w wood) FromString(s string) (interface{}, error) {
	switch s {
	case "oak":
		return Wood{wood(0)}, nil
	case "spruce":
		return Wood{wood(1)}, nil
	case "birch":
		return Wood{wood(2)}, nil
	case "jungle":
		return Wood{wood(3)}, nil
	case "acacia":
		return Wood{wood(4)}, nil
	case "dark_oak":
		return Wood{wood(5)}, nil
	}
	return nil, fmt.Errorf("unexpected wood type '%v', expecting one of 'oak', 'spruce', 'birch', 'jungle', 'acacia' or 'dark_oak'", s)
}

// String ...
func (w wood) String() string {
	switch w {
	case 0:
		return "oak"
	case 1:
		return "spruce"
	case 2:
		return "birch"
	case 3:
		return "jungle"
	case 4:
		return "acacia"
	case 5:
		return "dark_oak"
	}
	panic("unknown wood type")
}
