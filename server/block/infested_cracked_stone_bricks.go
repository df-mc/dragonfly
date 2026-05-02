package block

// InfestedCrackedStoneBricks is a block that hides a silverfish. It looks identical to cracked stone bricks.
type InfestedCrackedStoneBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedCrackedStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedCrackedStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_cracked_stone_bricks", 0
}

// EncodeBlock ...
func (i InfestedCrackedStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_cracked_stone_bricks", nil
}

// Hash ...
func (i InfestedCrackedStoneBricks) Hash() (uint64, uint64) {
	return hashInfestedCrackedStoneBricks, 0
}
