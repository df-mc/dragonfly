package block

// DriedKelp is a block primarily used as fuel in furnaces.
type DriedKelp struct {
	solid
}

// FlammabilityInfo ...
func (DriedKelp) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, false)
}

// BreakInfo ...
func (d DriedKelp) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, hoeEffective, oneOf(d))
}

// EncodeItem ...
func (DriedKelp) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp_block", 0
}

// EncodeBlock ...
func (DriedKelp) EncodeBlock() (string, map[string]any) {
	return "minecraft:dried_kelp_block", nil
}
