package item

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
	return Colours()[(oc - 1)], true
}

// Empty returns an empty OptionalColour.
func (oc OptionalColour) Empty() OptionalColour {
	return 0
}

// Uint8 returns the OptionalColour as an uint8.
func (oc OptionalColour) Uint8() uint8 {
	return uint8(oc)
}

// OptionalColours returns a slice of all possible OptionalColours.
func OptionalColours() []OptionalColour {
	colours := make([]OptionalColour, 17)
	for i, c := range Colours() {
		colours[i+1] = NewOptionalColour(c)
	}
	return colours
}
