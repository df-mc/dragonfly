package block

// ResinBricks is a block crafted from resin brick.
type ResinBricks struct {
	solid
	bassDrum

	// Chiseled specifies if the resin bricks is its chiseled variant.
	Chiseled bool
}

// BreakInfo ...
func (r ResinBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(30)
}

// EncodeItem ...
func (r ResinBricks) EncodeItem() (name string, meta int16) {
	if r.Chiseled {
		return "minecraft:chiseled_resin_bricks", 0
	}
	return "minecraft:resin_bricks", 0
}

// EncodeBlock ...
func (r ResinBricks) EncodeBlock() (string, map[string]any) {
	if r.Chiseled {
		return "minecraft:chiseled_resin_bricks", nil
	}
	return "minecraft:resin_bricks", nil
}
