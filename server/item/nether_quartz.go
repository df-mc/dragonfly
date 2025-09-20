package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// NetherQuartz is a smooth, white mineral found in the Nether.
type NetherQuartz struct{}

func (NetherQuartz) EncodeItem() (name string, meta int16) {
	return "minecraft:quartz", 0
}

func (NetherQuartz) TrimMaterial() string {
	return "quartz"
}

func (NetherQuartz) MaterialColour() string {
	return text.Quartz
}
