package block

// CoalOre is a common ore.
type CoalOre struct {
	solid
	bassDrum

	// Type is the type of coal ore.
	Type OreType
}

// BreakInfo ...
func (c CoalOre) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	b := newBreakInfo(c.Type.Hardness(), pickaxeHarvestable, pickaxeEffective, oneOf(c))
	b.XPDrops = XPDropRange{0, 2}
	return b
}

// EncodeItem ...
func (c CoalOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Type.Prefix() + "coal_ore", 0
}

// EncodeBlock ...
func (c CoalOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:" + c.Type.Prefix() + "coal_ore", nil

}
