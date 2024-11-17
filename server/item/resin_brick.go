package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// ResinBrick ...
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
