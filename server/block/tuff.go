package block

// Tuff is an ornamental rock formed from volcanic ash, occurring in underground blobs below Y=16.
type Tuff struct {
	solid
	bassDrum

	// Chiseled specifies if the tuff is chiseled.
	Chiseled bool
}

func (t Tuff) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(t)).withBlastResistance(30)
}

func (t Tuff) EncodeItem() (name string, meta int16) {
	if t.Chiseled {
		return "minecraft:chiseled_tuff", 0
	}
	return "minecraft:tuff", 0
}

func (t Tuff) EncodeBlock() (string, map[string]any) {
	if t.Chiseled {
		return "minecraft:chiseled_tuff", nil
	}
	return "minecraft:tuff", nil
}
