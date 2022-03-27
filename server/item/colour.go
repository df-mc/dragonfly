package item

import (
	"image/color"
)

// Colour represents the colour of a block. Typically, Minecraft blocks have a total of 16 different colours.
type Colour struct {
	colour
}

// ColourWhite returns the white colour.
func ColourWhite() Colour {
	return Colour{0}
}

// ColourOrange returns the orange colour.
func ColourOrange() Colour {
	return Colour{1}
}

// ColourMagenta returns the magenta colour.
func ColourMagenta() Colour {
	return Colour{2}
}

// ColourLightBlue returns the light blue colour.
func ColourLightBlue() Colour {
	return Colour{3}
}

// ColourYellow returns the yellow colour.
func ColourYellow() Colour {
	return Colour{4}
}

// ColourLime returns the lime colour.
func ColourLime() Colour {
	return Colour{5}
}

// ColourPink returns the pink colour.
func ColourPink() Colour {
	return Colour{6}
}

// ColourGrey returns the grey colour.
func ColourGrey() Colour {
	return Colour{7}
}

// ColourLightGrey returns the light grey colour.
func ColourLightGrey() Colour {
	return Colour{8}
}

// ColourCyan returns the cyan colour.
func ColourCyan() Colour {
	return Colour{9}
}

// ColourPurple returns the purple colour.
func ColourPurple() Colour {
	return Colour{10}
}

// ColourBlue returns the blue colour.
func ColourBlue() Colour {
	return Colour{11}
}

// ColourBrown returns the brown colour.
func ColourBrown() Colour {
	return Colour{12}
}

// ColourGreen returns the green colour.
func ColourGreen() Colour {
	return Colour{13}
}

// ColourRed returns the red colour.
func ColourRed() Colour {
	return Colour{14}
}

// ColourBlack returns the black colour.
func ColourBlack() Colour {
	return Colour{15}
}

// Colours returns a list of all existing colours.
func Colours() []Colour {
	return []Colour{
		ColourWhite(), ColourOrange(), ColourMagenta(), ColourLightBlue(), ColourYellow(), ColourLime(), ColourPink(), ColourGrey(),
		ColourLightGrey(), ColourCyan(), ColourPurple(), ColourBlue(), ColourBrown(), ColourGreen(), ColourRed(), ColourBlack(),
	}
}

// colour is the underlying value of a Colour struct.
type colour uint8

// RGBA returns the colour as RGBA. The alpha channel is always set to the maximum value. Colour values as returned here
// were obtained by placing signs in a world with all possible dyes used on them. The world was then loaded in Dragonfly
// to read their respective colours.
func (c colour) RGBA() color.RGBA {
	switch c {
	case 0:
		return color.RGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}
	case 1:
		return color.RGBA{R: 0xf9, G: 0x80, B: 0x1d, A: 0xff}
	case 2:
		return color.RGBA{R: 0xc7, G: 0x4e, B: 0xbd, A: 0xff}
	case 3:
		return color.RGBA{R: 0x3a, G: 0xb3, B: 0xda, A: 0xff}
	case 4:
		return color.RGBA{R: 0xfe, G: 0xd8, B: 0x3d, A: 0xff}
	case 5:
		return color.RGBA{R: 0x80, G: 0xc7, B: 0x1f, A: 0xff}
	case 6:
		return color.RGBA{R: 0xf3, G: 0x8b, B: 0xaa, A: 0xff}
	case 7:
		return color.RGBA{R: 0x47, G: 0x4f, B: 0x52, A: 0xff}
	case 8:
		return color.RGBA{R: 0x9d, G: 0x9d, B: 0x97, A: 0xff}
	case 9:
		return color.RGBA{R: 0x16, G: 0x9c, B: 0x9c, A: 0xff}
	case 10:
		return color.RGBA{R: 0x89, G: 0x32, B: 0xb8, A: 0xff}
	case 11:
		return color.RGBA{R: 0x3c, G: 0x44, B: 0xaa, A: 0xff}
	case 12:
		return color.RGBA{R: 0x83, G: 0x54, B: 0x32, A: 0xff}
	case 13:
		return color.RGBA{R: 0x5e, G: 0x7c, B: 0x16, A: 0xff}
	case 14:
		return color.RGBA{R: 0xb0, G: 0x2e, B: 0x26, A: 0xff}
	default:
		return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	}
}

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
