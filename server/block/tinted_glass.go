package block

// TintedGlass is a decorative, solid block that is visually see-through but, unlike regular glass, blocks
// all light passing through it.
type TintedGlass struct {
	solid
	clicksAndSticks
}

// BreakInfo ...
func (g TintedGlass) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(g)).withBlastResistance(0.3)
}

// EncodeItem ...
func (TintedGlass) EncodeItem() (name string, meta int16) {
	return "minecraft:tinted_glass", 0
}

// EncodeBlock ...
func (TintedGlass) EncodeBlock() (string, map[string]any) {
	return "minecraft:tinted_glass", nil
}
