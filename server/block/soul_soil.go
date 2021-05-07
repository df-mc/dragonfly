package block

// SoulSoil is a block naturally found only in the soul sand valley.
type SoulSoil struct {
	solid
}

// BreakInfo ...
func (s SoulSoil) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

// EncodeItem ...
func (SoulSoil) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_soil", 0
}

// EncodeBlock ...
func (SoulSoil) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:soul_soil", nil
}
