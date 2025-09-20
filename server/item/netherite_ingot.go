package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// NetheriteIngot is a rare mineral crafted with 4 pieces of netherite scrap and 4 gold ingots.
type NetheriteIngot struct{}

func (NetheriteIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:netherite_ingot", 0
}

func (NetheriteIngot) TrimMaterial() string {
	return "netherite"
}

func (NetheriteIngot) MaterialColour() string {
	return text.Netherite
}

func (NetheriteIngot) PayableForBeacon() bool {
	return true
}
