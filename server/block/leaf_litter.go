package block

// LeafLitter is a bioluminescent block found abundantly in the deep dark
type LeafLitter struct {
	solid
}

// BreakInfo ...
func (l LeafLitter) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, hoeEffective, silkTouchOnlyDrop(l)).withXPDropRange(1, 1)
}

// EncodeItem ...
func (l LeafLitter) EncodeItem() (name string, meta int16) {
	return "minecraft:leaf_litter", 0
}

// EncodeBlock ...
func (l LeafLitter) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:leaf_litter", nil
}
