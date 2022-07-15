package block

// Cobblestone is a common block, obtained from mining stone.
type Cobblestone struct {
	solid
	bassDrum

	// Mossy specifies if the cobblestone is mossy. This variant of cobblestone is typically found in
	// dungeons or in small clusters in the giant tree taiga biome.
	Mossy bool
}

// BreakInfo ...
func (c Cobblestone) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

// RepairsStoneTools ...
func (c Cobblestone) RepairsStoneTools() bool {
	return !c.Mossy
}

// EncodeItem ...
func (c Cobblestone) EncodeItem() (name string, meta int16) {
	if c.Mossy {
		return "minecraft:mossy_cobblestone", 0
	}
	return "minecraft:cobblestone", 0
}

// EncodeBlock ...
func (c Cobblestone) EncodeBlock() (string, map[string]any) {
	if c.Mossy {
		return "minecraft:mossy_cobblestone", nil
	}
	return "minecraft:cobblestone", nil
}
