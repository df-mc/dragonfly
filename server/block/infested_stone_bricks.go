package block

// InfestedStoneBricks is a block that hides a silverfish. It looks identical to stone bricks.
type InfestedStoneBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_stone_bricks", 0
}

// EncodeBlock ...
func (i InfestedStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_stone_bricks", nil
}

// Hash ...
func (i InfestedStoneBricks) Hash() (uint64, uint64) {
	return hashInfestedStoneBricks, 0
}
