package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// AmethystShard is a crystalline mineral obtained from mining a fully grown amethyst cluster.
type AmethystShard struct{}

// EncodeItem ...
func (AmethystShard) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_shard", 0
}

// TrimMaterial ...
func (AmethystShard) TrimMaterial() string {
	return "amethyst"
}

// MaterialColour ...
func (AmethystShard) MaterialColour() string {
	return text.Amethyst
}
