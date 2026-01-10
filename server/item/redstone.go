package item

// Redstone is a resource obtained from mining redstone ore.
type Redstone struct{}

// EncodeItem ...
func (Redstone) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}
