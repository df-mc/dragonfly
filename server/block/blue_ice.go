package block

// BlueIce is a solid block similar to packed ice.
type BlueIce struct {
	solid
}

func (b BlueIce) BreakInfo() BreakInfo {
	return newBreakInfo(2.8, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(b))
}

func (b BlueIce) Friction() float64 {
	return 0.989
}

func (BlueIce) EncodeItem() (name string, meta int16) {
	return "minecraft:blue_ice", 0
}

func (BlueIce) EncodeBlock() (string, map[string]any) {
	return "minecraft:blue_ice", nil
}
