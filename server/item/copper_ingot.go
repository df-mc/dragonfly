package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// CopperIngot is a metal ingot melted from copper ore.
type CopperIngot struct{}

// EncodeItem ...
func (c CopperIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:copper_ingot", 0
}

// TrimMaterial ...
func (CopperIngot) TrimMaterial() string {
	return "copper"
}

// MaterialColour ...
func (CopperIngot) MaterialColour() string {
	return text.Copper
}
