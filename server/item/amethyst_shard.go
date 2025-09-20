package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// AmethystShard is a crystalline mineral obtained from mining a fully grown amethyst cluster.
type AmethystShard struct{}

func (AmethystShard) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_shard", 0
}

func (AmethystShard) TrimMaterial() string {
	return "amethyst"
}

func (AmethystShard) MaterialColour() string {
	return text.Amethyst
}
