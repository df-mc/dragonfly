package colour

// Colour represents the colour of a block. Typically, Minecraft blocks have a total of 16 different colours.
type Colour interface {
	// String returns a string that represents the colour in Minecraft, such as 'light_gray'.
	String() string
	// Uint8 returns a uint8 that represents the colour in Minecraft as an int.
	Uint8() uint8
	__()
}

// White returns the white colour.
func White() Colour {
	return colour(0)
}

// Orange returns the orange colour.
func Orange() Colour {
	return colour(1)
}

// Magenta returns the magenta colour.
func Magenta() Colour {
	return colour(2)
}

// LightBlue returns the light blue colour.
func LightBlue() Colour {
	return colour(3)
}

// Yellow returns the yellow colour.
func Yellow() Colour {
	return colour(4)
}

// Lime returns the lime colour.
func Lime() Colour {
	return colour(5)
}

// Pink returns the pink colour.
func Pink() Colour {
	return colour(6)
}

// Grey returns the grey colour.
func Grey() Colour {
	return colour(7)
}

// LightGrey returns the light grey colour.
func LightGrey() Colour {
	return colour(8)
}

// Cyan returns the cyan colour.
func Cyan() Colour {
	return colour(9)
}

// Purple returns the purple colour.
func Purple() Colour {
	return colour(10)
}

// Blue returns the blue colour.
func Blue() Colour {
	return colour(11)
}

// Brown returns the brown colour.
func Brown() Colour {
	return colour(12)
}

// Green returns the green colour.
func Green() Colour {
	return colour(13)
}

// Red returns the red colour.
func Red() Colour {
	return colour(14)
}

// Black returns the black colour.
func Black() Colour {
	return colour(15)
}

// All returns a list of all existing colours.
func All() []Colour {
	return []Colour{
		White(), Orange(), Magenta(), LightBlue(), Yellow(), Lime(), Pink(), Grey(),
		LightGrey(), Cyan(), Purple(), Blue(), Brown(), Green(), Red(), Black(),
	}
}

type colour uint8

func (colour) __() {}

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

// Uint8 ...
func (c colour) Uint8() uint8 {
	return uint8(c)
}
