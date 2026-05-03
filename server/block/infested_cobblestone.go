package block

// InfestedCobblestone is a block that hides a silverfish. It looks identical to cobblestone.
type InfestedCobblestone struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedCobblestone) BreakInfo() BreakInfo {
	return newBreakInfo(1, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedCobblestone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_cobblestone", 0
}

// EncodeBlock ...
func (i InfestedCobblestone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_cobblestone", nil
}
