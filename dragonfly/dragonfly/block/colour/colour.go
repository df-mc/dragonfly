package colour

import "fmt"

// Colour represents the colour of a block. Typically, Minecraft blocks have a total of 16 different colours.
type Colour struct {
	colour
}

// White returns the white colour.
func White() Colour {
	return Colour{colour(0)}
}

// Orange returns the orange colour.
func Orange() Colour {
	return Colour{colour(1)}
}

// Magenta returns the magenta colour.
func Magenta() Colour {
	return Colour{colour(2)}
}

// LightBlue returns the light blue colour.
func LightBlue() Colour {
	return Colour{colour(3)}
}

// Yellow returns the yellow colour.
func Yellow() Colour {
	return Colour{colour(4)}
}

// Lime returns the lime colour.
func Lime() Colour {
	return Colour{colour(5)}
}

// Pink returns the pink colour.
func Pink() Colour {
	return Colour{colour(6)}
}

// Grey returns the grey colour.
func Grey() Colour {
	return Colour{colour(7)}
}

// LightGrey returns the light grey colour.
func LightGrey() Colour {
	return Colour{colour(8)}
}

// Cyan returns the cyan colour.
func Cyan() Colour {
	return Colour{colour(9)}
}

// Purple returns the purple colour.
func Purple() Colour {
	return Colour{colour(10)}
}

// Blue returns the blue colour.
func Blue() Colour {
	return Colour{colour(11)}
}

// Brown returns the brown colour.
func Brown() Colour {
	return Colour{colour(12)}
}

// Green returns the green colour.
func Green() Colour {
	return Colour{colour(13)}
}

// Red returns the red colour.
func Red() Colour {
	return Colour{colour(14)}
}

// Black returns the black colour.
func Black() Colour {
	return Colour{colour(15)}
}

// All returns a list of all existing colours.
func All() []Colour {
	return []Colour{
		White(), Orange(), Magenta(), LightBlue(), Yellow(), Lime(), Pink(), Grey(),
		LightGrey(), Cyan(), Purple(), Blue(), Brown(), Green(), Red(), Black(),
	}
}

type colour uint8

// String ...
func (c colour) String() string {
	switch c {
	default:
		return "white"
	case 1:
		return "orange"
	case 2:
		return "magenta"
	case 3:
		return "light_blue"
	case 4:
		return "yellow"
	case 5:
		return "lime"
	case 6:
		return "pink"
	case 7:
		return "gray"
	case 8:
		return "silver"
	case 9:
		return "cyan"
	case 10:
		return "purple"
	case 11:
		return "blue"
	case 12:
		return "brown"
	case 13:
		return "green"
	case 14:
		return "red"
	case 15:
		return "black"
	}
}

// FromString ...
func (c colour) FromString(s string) (interface{}, error) {
	switch s {
	case "white":
		return Colour{colour(0)}, nil
	case "orange":
		return Colour{colour(1)}, nil
	case "magenta":
		return Colour{colour(2)}, nil
	case "light_blue":
		return Colour{colour(3)}, nil
	case "yellow":
		return Colour{colour(4)}, nil
	case "lime", "light_green":
		return Colour{colour(5)}, nil
	case "pink":
		return Colour{colour(6)}, nil
	case "grey", "gray":
		return Colour{colour(7)}, nil
	case "light_grey", "light_gray", "silver":
		return Colour{colour(8)}, nil
	case "cyan":
		return Colour{colour(9)}, nil
	case "purple":
		return Colour{colour(10)}, nil
	case "blue":
		return Colour{colour(11)}, nil
	case "brown":
		return Colour{colour(12)}, nil
	case "green":
		return Colour{colour(13)}, nil
	case "red":
		return Colour{colour(14)}, nil
	case "black":
		return Colour{colour(15)}, nil
	}
	return nil, fmt.Errorf("unexpected colour '%v'", s)
}

// Uint8 ...
func (c colour) Uint8() uint8 {
	return uint8(c)
}
