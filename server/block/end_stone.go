package block

// EndStone is a block found in The End.
type EndStone struct {
	solid
	bassDrum
}

func (e EndStone) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(e)).withBlastResistance(45)
}

func (EndStone) EncodeItem() (name string, meta int16) {
	return "minecraft:end_stone", 0
}

func (EndStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_stone", nil
}
