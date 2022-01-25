package block

// Slime is a full-sized block of snow.
type Slime struct {
	solid
}

// BreakInfo ...
func (s Slime) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, shovelEffective, oneOf(s))
}

// EncodeItem ...
func (Slime) EncodeItem() (name string, meta int16) {
	return "minecraft:slime", 0
}

// EncodeBlock ...
func (Slime) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:slime", nil
}
