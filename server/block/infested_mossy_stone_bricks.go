package block

// InfestedMossyStoneBricks is a block that hides a silverfish. It looks identical to mossy stone bricks.
type InfestedMossyStoneBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedMossyStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedMossyStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_mossy_stone_bricks", 0
}

// EncodeBlock ...
func (i InfestedMossyStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_mossy_stone_bricks", nil
}
