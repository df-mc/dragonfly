package block

// Glass is a decorative, fully transparent solid block that can be dyed into stained-glass.
type Glass struct {
	solid
	transparent
	clicksAndSticks
}

func (g Glass) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, silkTouchOnlyDrop(g))
}

func (Glass) EncodeItem() (name string, meta int16) {
	return "minecraft:glass", 0
}

func (Glass) EncodeBlock() (string, map[string]any) {
	return "minecraft:glass", nil
}
