package block

// TODO: Slipperiness

// BlueIce is a solid block similar to packed ice.
type BlueIce struct {
	noNBT
	solid
}

// LightEmissionLevel ...
func (BlueIce) LightEmissionLevel() uint8 {
	return 4
}

// BreakInfo ...
func (b BlueIce) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2.8,
		Harvestable: alwaysHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(),
	}
}

// EncodeItem ...
func (BlueIce) EncodeItem() (id int32, meta int16) {
	return -11, 0
}
