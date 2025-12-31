package block

import "github.com/df-mc/dragonfly/server/item"

// OptionalColour is an optional colour for certain blocks.
type OptionalColour uint8

// NewOptionalColour returns a new OptionalColour from an item.Colour.
func NewOptionalColour(c item.Colour) OptionalColour {
	return OptionalColour(c.Uint8() + 1)
}

// Colour returns the item.Colour of the OptionalColour.
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

// Uint8 returns the OptionalColour as an uint8.
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
