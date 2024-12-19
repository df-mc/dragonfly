package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// ResinBrick is an item used to create resin bricks. It can also be used as a
// smithing ingredient, giving orange details to pieces of armor.
type ResinBrick struct{}

// EncodeItem ...
func (ResinBrick) EncodeItem() (name string, meta int16) {
	return "minecraft:resin_brick", 0
}

// TrimMaterial ...
func (ResinBrick) TrimMaterial() string {
	return "resin"
}

// MaterialColour ...
func (ResinBrick) MaterialColour() string {
	return text.Resin
}
