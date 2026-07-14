package block

// CinnabarBricks is a decorative variant of Cinnabar.
type CinnabarBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (c CinnabarBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(6)
}

// EncodeItem ...
func (CinnabarBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:cinnabar_bricks", 0
}

// EncodeBlock ...
func (CinnabarBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:cinnabar_bricks", nil
}
