package block

// InfestedStone is a block that hides a silverfish. It looks identical to stone.
// TODO: spawn a silverfish on break (without silk touch) once silverfish are implemented.
type InfestedStone struct {
	solid
	flute
}

// BreakInfo ...
func (i InfestedStone) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBlastResistance(0.75)
}

// EncodeItem ...
func (InfestedStone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_stone", 0
}

// EncodeBlock ...
func (InfestedStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_stone", nil
}
