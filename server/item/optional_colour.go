package item

// OptionalColour represents a Colour that may be absent. It is used by items
// and blocks that have one variant per Colour plus an additional uncoloured
// variant (e.g. shulker boxes, where the uncoloured form is encoded as
// minecraft:undyed_shulker_box). A zero OptionalColour denotes the absent
// case; values 1..16 map to the 16 dye Colours.
type OptionalColour uint8

// colours is a slice of all Colours.
var colours = Colours()

// NewOptionalColour returns a new OptionalColour from a Colour.
func NewOptionalColour(c Colour) OptionalColour {
	return OptionalColour(c.Uint8() + 1)
}

// Colour returns the Colour of the OptionalColour, and a bool indicating
// whether the Colour is present.
func (oc OptionalColour) Colour() (Colour, bool) {
	if oc == 0 {
		return Colour{}, false
	}
	return colours[(oc - 1)], true
}

// Uint8 returns the OptionalColour as a uint8.
func (oc OptionalColour) Uint8() uint8 {
	return uint8(oc)
}

// Prepend prepends the Colour to the string if the Colour is present.
func (oc OptionalColour) Prepend(str string) string {
	if oc != 0 {
		return colours[(oc-1)].String() + "_" + str
	}
	return str
}

// OptionalColours returns a slice of all OptionalColours, including the absent case.
func OptionalColours() []OptionalColour {
	optionalColours := make([]OptionalColour, 17)
	for i, c := range colours {
		optionalColours[i+1] = NewOptionalColour(c)
	}
	return optionalColours
}
