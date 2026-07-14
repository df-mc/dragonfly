package block

// SulfurBricks is a decorative variant of Sulfur.
type SulfurBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (s SulfurBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(6)
}

// EncodeItem ...
func (SulfurBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:sulfur_bricks", 0
}

// EncodeBlock ...
func (SulfurBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:sulfur_bricks", nil
}
