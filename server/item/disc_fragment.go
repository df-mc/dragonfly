package item

// DiscFragment is a music disc fragment obtained from ancient city loot chests. They are extremely rare to find and
// nine of them in a crafting table makes a music disc named, "5".
type DiscFragment struct{}

// EncodeItem ...
func (DiscFragment) EncodeItem() (name string, meta int16) {
	return "minecraft:disc_fragment_5", 0
}
