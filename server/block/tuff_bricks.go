package block

// TuffBricks are a decorational variant of Tuff that can be crafted or found naturally in Trial Chambers.
type TuffBricks struct {
	solid
	bassDrum

	// Chiseled specifies if the tuff bricks are chiseled.
	Chiseled bool
}

// BreakInfo ...
func (t TuffBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(t)).withBlastResistance(30)
}

// EncodeItem ...
func (t TuffBricks) EncodeItem() (name string, meta int16) {
	if t.Chiseled {
		return "minecraft:chiseled_tuff_bricks", 0
	}
	return "minecraft:tuff_bricks", 0
}

// EncodeBlock ...
func (t TuffBricks) EncodeBlock() (string, map[string]any) {
	if t.Chiseled {
		return "minecraft:chiseled_tuff_bricks", nil
	}
	return "minecraft:tuff_bricks", nil
}
