package block

// Sculk is a bioluminescent block found abundantly in the deep dark
type Sculk struct {
	solid
}

// BreakInfo ...
func (s Sculk) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, hoeEffective, silkTouchOnlyDrop(s)).withXPDropRange(1, 1)
}

// EncodeItem ...
func (s Sculk) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk", 0
}

// EncodeBlock ...
func (s Sculk) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:sculk", nil
}
