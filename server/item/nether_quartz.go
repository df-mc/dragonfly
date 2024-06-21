package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// NetherQuartz is a smooth, white mineral found in the Nether.
type NetherQuartz struct{}

// EncodeItem ...
func (NetherQuartz) EncodeItem() (name string, meta int16) {
	return "minecraft:quartz", 0
}

// TrimMaterial ...
func (NetherQuartz) TrimMaterial() string {
	return "quartz"
}

// MaterialColour ...
func (NetherQuartz) MaterialColour() string {
	return text.Quartz
}
