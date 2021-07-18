package block

// TODO: Slipperiness

// BlueIce is a solid block similar to packed ice.
type BlueIce struct {
	solid
}

// BreakInfo ...
func (b BlueIce) BreakInfo() BreakInfo {
	return newBreakInfo(2.8, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(b))
}

// EncodeItem ...
func (BlueIce) EncodeItem() (name string, meta int16) {
	return "minecraft:blue_ice", 0
}

// EncodeBlock ...
func (BlueIce) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:blue_ice", nil
}
