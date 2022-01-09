package block

// DriedKelpBlock is a block primarily used as fuel in furnaces.
type DriedKelpBlock struct {
	solid
}

// FlammabilityInfo ...
func (DriedKelpBlock) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, false)
}

// BreakInfo ...
func (d DriedKelpBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, hoeEffective, oneOf(d))
}

// EncodeItem ...
func (DriedKelpBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp_block", 0
}

// EncodeBlock ...
func (DriedKelpBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:dried_kelp_block", nil
}
