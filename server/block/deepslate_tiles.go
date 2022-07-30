package block

// DeepslateTiles are a tiled variant of deepslate and can spawn in ancient cities.
type DeepslateTiles struct {
	solid
	bassDrum

	// Cracked specifies if the deepslate tiles is its cracked variant.
	Cracked bool
}

// BreakInfo ...
func (d DeepslateTiles) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(18)
}

// EncodeItem ...
func (d DeepslateTiles) EncodeItem() (name string, meta int16) {
	if d.Cracked {
		return "minecraft:cracked_deepslate_tiles", 0
	}
	return "minecraft:deepslate_tiles", 0
}

// EncodeBlock ...
func (d DeepslateTiles) EncodeBlock() (string, map[string]any) {
	if d.Cracked {
		return "minecraft:cracked_deepslate_tiles", nil
	}
	return "minecraft:deepslate_tiles", nil
}
