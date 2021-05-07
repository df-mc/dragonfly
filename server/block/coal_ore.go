package block

// CoalOre is a common ore.
type CoalOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (c CoalOre) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	b := newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(c))
	b.XPDrops = XPDropRange{0, 2}
	return b
}

// EncodeItem ...
func (CoalOre) EncodeItem() (name string, meta int16) {
	return "minecraft:coal_ore", 0
}

// EncodeBlock ...
func (CoalOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:coal_ore", nil
}
