package block

// Resin is a block equivalent to nine resin clumps.
type Resin struct {
	solid
}

func (r Resin) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r))
}

func (Resin) EncodeItem() (name string, meta int16) {
	return "minecraft:resin_block", 0
}

func (Resin) EncodeBlock() (string, map[string]any) {
	return "minecraft:resin_block", nil
}
