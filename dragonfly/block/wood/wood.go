package wood

import "fmt"

// Wood represents a type of wood of a block. Some blocks, such as log blocks, bark blocks, wooden planks and
// others carry one of these types.
type Wood struct {
	wood
	Flammable bool
}

// Oak returns oak wood material.
func Oak() Wood {
	return Wood{wood(0), true}
}

// Spruce returns spruce wood material.
func Spruce() Wood {
	return Wood{wood(1), true}
}

// Birch returns birch wood material.
func Birch() Wood {
	return Wood{wood(2), true}
}

// Jungle returns jungle wood material.
func Jungle() Wood {
	return Wood{wood(3), true}
}

// Acacia returns acacia wood material.
func Acacia() Wood {
	return Wood{wood(4), true}
}

// DarkOak returns dark oak wood material.
func DarkOak() Wood {
	return Wood{wood(5), true}
}

// Crimson returns crimson wood material.
func Crimson() Wood {
	return Wood{wood(6), false}
}

// Warped returns warped wood material.
func Warped() Wood {
	return Wood{wood(7), false}
}

// All returns a list of all wood types
func All() []Wood {
	return []Wood{Oak(), Spruce(), Birch(), Jungle(), Acacia(), DarkOak(), Crimson(), Warped()}
}

type wood uint8

// Uint8 returns the wood as a uint8.
func (w wood) Uint8() uint8 {
	return uint8(w)
}

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
	case 6:
		return "Crimson"
	case 7:
		return "Warped"
	}
	panic("unknown wood type")
}

// FromString ...
func (w wood) FromString(s string) (interface{}, error) {
	switch s {
	case "oak":
		return Wood{wood(0), true}, nil
	case "spruce":
		return Wood{wood(1), true}, nil
	case "birch":
		return Wood{wood(2), true}, nil
	case "jungle":
		return Wood{wood(3), true}, nil
	case "acacia":
		return Wood{wood(4), true}, nil
	case "dark_oak":
		return Wood{wood(5), true}, nil
	case "crimson":
		return Wood{wood(6), false}, nil
	case "warped":
		return Wood{wood(7), false}, nil
	}
	return nil, fmt.Errorf("unexpected wood type '%v', expecting one of 'oak', 'spruce', 'birch', 'jungle', 'acacia', 'dark_oak', 'crimson' or 'warped'", s)
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
	case 6:
		return "crimson"
	case 7:
		return "warped"
	}
	panic("unknown wood type")
}
