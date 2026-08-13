package block

// PolishedCinnabar is a decorative variant of Cinnabar.
type PolishedCinnabar struct {
	solid
	bassDrum
}

// BreakInfo ...
func (c PolishedCinnabar) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(6)
}

// EncodeItem ...
func (PolishedCinnabar) EncodeItem() (name string, meta int16) {
	return "minecraft:polished_cinnabar", 0
}

// EncodeBlock ...
func (PolishedCinnabar) EncodeBlock() (string, map[string]any) {
	return "minecraft:polished_cinnabar", nil
}
