package item

import "github.com/df-mc/dragonfly/server/item/potion"

// Arrow is used as ammunition for bows, crossbows, and dispensers. Arrows can be modified to
// imbue status effects on players and mobs.
type Arrow struct {
	// Tip is the potion effect that is tipped on the arrow.
	Tip potion.Potion
}

// EncodeItem ...
func (a Arrow) EncodeItem() (name string, meta int16) {
	if tip := a.Tip.Uint8(); tip > 4 {
		return "minecraft:arrow", int16(tip + 1)
	}
	return "minecraft:arrow", 0
}
