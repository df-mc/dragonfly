package item

// colours is a slice of all Colours.
var colours = Colours()

// OptionalColour is an optional Colour for certain items and blocks.
type OptionalColour uint8

// NewOptionalColour returns a new OptionalColour from a Colour.
func NewOptionalColour(c Colour) OptionalColour {
	return OptionalColour(c.Uint8() + 1)
}

// Colour returns the Colour of the OptionalColour.
func (oc OptionalColour) Colour() (Colour, bool) {
	if oc == 0 {
		return Colour{}, false
	}
	return colours[(oc - 1)], true
}

// Uint8 returns the OptionalColour as an uint8.
func (oc OptionalColour) Uint8() uint8 {
	return uint8(oc)
}

// Prepend prepends the Colour to the string if the Colour is not 0.
func (oc OptionalColour) Prepend(str string) string {
	if oc != 0 {
		return colours[(oc-1)].String() + "_" + str
	}
	return str
}

// OptionalColours returns a slice of all possible OptionalColours.
func OptionalColours() []OptionalColour {
	optionalColours := make([]OptionalColour, 17)
	for i, c := range colours {
		optionalColours[i+1] = NewOptionalColour(c)
	}
	return optionalColours
}
