package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// GoldIngot is a metal ingot melted from raw gold or obtained from loot chests.
type GoldIngot struct{}

// EncodeItem ...
func (GoldIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:gold_ingot", 0
}

// TrimMaterial ...
func (GoldIngot) TrimMaterial() string {
	return "gold"
}

// MaterialColour ...
func (GoldIngot) MaterialColour() string {
	return text.Gold
}

// PayableForBeacon ...
func (GoldIngot) PayableForBeacon() bool {
	return true
}
