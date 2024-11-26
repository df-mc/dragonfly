package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// Diamond is a rare mineral obtained from diamond ore or loot chests.
type Diamond struct{}

// EncodeItem ...
func (Diamond) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond", 0
}

// TrimMaterial ...
func (Diamond) TrimMaterial() string {
	return "diamond"
}

// MaterialColour ...
func (Diamond) MaterialColour() string {
	return text.Diamond
}

// PayableForBeacon ...
func (Diamond) PayableForBeacon() bool {
	return true
}
