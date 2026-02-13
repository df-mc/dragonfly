package block

// SmoothBasalt is a decorative solid block obtained by smelting basalt.
type SmoothBasalt struct {
	solid
	bassDrum
}

// EncodeBlock ...
func (SmoothBasalt) EncodeBlock() (string, map[string]any) {
	return "minecraft:smooth_basalt", nil
}

// EncodeItem ...
func (SmoothBasalt) EncodeItem() (name string, meta int16) {
	return "minecraft:smooth_basalt", 0
}

// BreakInfo ...
func (s SmoothBasalt) BreakInfo() BreakInfo {
	return newBreakInfo(1.25, pickaxeHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(21)
}
