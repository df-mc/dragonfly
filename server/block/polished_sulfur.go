package block

// PolishedSulfur is a decorative variant of Sulfur.
type PolishedSulfur struct {
	solid
	bassDrum
}

// BreakInfo ...
func (s PolishedSulfur) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(6)
}

// EncodeItem ...
func (PolishedSulfur) EncodeItem() (name string, meta int16) {
	return "minecraft:polished_sulfur", 0
}

// EncodeBlock ...
func (PolishedSulfur) EncodeBlock() (string, map[string]any) {
	return "minecraft:polished_sulfur", nil
}
