package block

// EndStone is a block found in The End.
type EndStone struct {
	solid
	bassDrum
}

// BreakInfo ...
func (e EndStone) BreakInfo() BreakInfo {
	return NewBreakInfo(3, PickaxeHarvestable, PickaxeEffective, OneOf(e)).withBlastResistance(45)
}

// EncodeItem ...
func (EndStone) EncodeItem() (name string, meta int16) {
	return "minecraft:end_stone", 0
}

// EncodeBlock ...
func (EndStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_stone", nil
}
