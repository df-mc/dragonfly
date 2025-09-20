package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// Diamond is a rare mineral obtained from diamond ore or loot chests.
type Diamond struct{}

func (Diamond) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond", 0
}

func (Diamond) TrimMaterial() string {
	return "diamond"
}

func (Diamond) MaterialColour() string {
	return text.Diamond
}

func (Diamond) PayableForBeacon() bool {
	return true
}
