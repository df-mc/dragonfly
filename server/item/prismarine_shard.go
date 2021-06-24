package item

// PrismarineShard is an item obtained by defeating guardians or elder guardians.
type PrismarineShard struct{}

// EncodeItem ...
func (PrismarineShard) EncodeItem() (name string, meta int16) {
	return "minecraft:prismarine_shard", 0
}
