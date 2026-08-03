package block

// Cinnabar is a decorative rock that generates throughout sulfur caves and as part of sulfur springs.
type Cinnabar struct {
	solid
	bassDrum

	// Chiseled specifies if the cinnabar is chiseled.
	Chiseled bool
}

// BreakInfo ...
func (c Cinnabar) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(6)
}

// EncodeItem ...
func (c Cinnabar) EncodeItem() (name string, meta int16) {
	if c.Chiseled {
		return "minecraft:chiseled_cinnabar", 0
	}
	return "minecraft:cinnabar", 0
}

// EncodeBlock ...
func (c Cinnabar) EncodeBlock() (string, map[string]any) {
	if c.Chiseled {
		return "minecraft:chiseled_cinnabar", nil
	}
	return "minecraft:cinnabar", nil
}
