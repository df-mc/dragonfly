package item

// AmethystShard is a crystalline mineral obtained from mining a fully grown amethyst cluster.
type AmethystShard struct{}

// EncodeItem ...
func (AmethystShard) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_shard", 0
}
