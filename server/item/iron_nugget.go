package item

// IronNugget is a piece of iron that can be obtained by smelting iron tools/weapons or iron/chainmail armor.
type IronNugget struct{}

// EncodeItem ...
func (IronNugget) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_nugget", 0
}
