package item

import "github.com/df-mc/dragonfly/server/world"

// Dyes are made up of 16 different type of colours which allows you to dye blocks like concrete and sheep.
type Dyes struct {
	Colour Colour
}

// AllDyes returns all 16 dye items
func AllDyes() []world.Item {
	b := make([]world.Item, 0, 16)
	for _, c := range Colours() {
		b = append(b, Dyes{Colour: c})
	}
	return b
}

// EncodeItem ...
func (d Dyes) EncodeItem() (name string, meta int16) {
	if d.Colour.String() == "silver" {
		return "minecraft:light_gray_dye", 0
	}
	return "minecraft:" + d.Colour.String() + "_dye", 0
}
