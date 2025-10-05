package item

// CopperNugget is a piece of copper that can be obtained by smelting copper tools/weapons or armour.
type CopperNugget struct{}

// EncodeItem ...
func (CopperNugget) EncodeItem() (name string, meta int16) {
	return "minecraft:copper_nugget", 0
}
