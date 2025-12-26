package block

import "github.com/df-mc/dragonfly/server/item"

// OptionalColour is an optional colour that can be used to represent a block's colour.'
type OptionalColour uint8

// NewOptionalColour creates a new OptionalColour from a Colour.
func NewOptionalColour(c item.Colour) OptionalColour {
	return OptionalColour(c.Uint8() + 1)
}

// Colour returns the colour of the OptionalColour.
func (oc OptionalColour) Colour() (item.Colour, bool) {
	if oc == 0 {
		return item.Colour{}, false
	}
	return item.Colours()[(oc - 1)], true
}

// Empty returns an empty OptionalColour.
func (oc OptionalColour) Empty() OptionalColour {
	return 0
}

// Uint8 returns the uint8 representation of the OptionalColour.
func (oc OptionalColour) Uint8() uint8 {
	return uint8(oc)
}

// OptionalColours returns a slice of all possible OptionalColours.
func OptionalColours() []OptionalColour {
	colours := make([]OptionalColour, 17)
	for i, c := range item.Colours() {
		colours[i+1] = NewOptionalColour(c)
	}
	return colours
}
