package block

// InfestedChiseledStoneBricks is a block that hides a silverfish. It looks identical to chiseled stone bricks.
type InfestedChiseledStoneBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedChiseledStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedChiseledStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_chiseled_stone_bricks", 0
}

// EncodeBlock ...
func (i InfestedChiseledStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_chiseled_stone_bricks", nil
}

// Hash ...
func (i InfestedChiseledStoneBricks) Hash() (uint64, uint64) {
	return hashInfestedChiseledStoneBricks, 0
}
