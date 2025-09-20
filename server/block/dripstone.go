package block

// Dripstone is a rock block that allows pointed dripstone to grow beneath it.
type Dripstone struct {
	solid
	bassDrum
}

func (d Dripstone) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(5)
}

func (d Dripstone) EncodeItem() (name string, meta int16) {
	return "minecraft:dripstone_block", 0
}

func (d Dripstone) EncodeBlock() (string, map[string]any) {
	return "minecraft:dripstone_block", nil
}
