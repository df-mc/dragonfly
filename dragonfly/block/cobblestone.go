package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"

// Cobblestone is a common block, obtained from mining stone.
type Cobblestone struct {
	// Mossy specifies if the cobblestone is mossy. This variant of cobblestone is typically found in
	// dungeons or in small clusters in the giant tree taiga biome.
	Mossy bool
}

// BreakInfo ...
func (c Cobblestone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(c, 1)),
	}
}

// EncodeItem ...
func (c Cobblestone) EncodeItem() (id int32, meta int16) {
	if c.Mossy {
		return 48, 0
	}
	return 4, 0
}

// EncodeBlock ...
func (c Cobblestone) EncodeBlock() (name string, properties map[string]interface{}) {
	if c.Mossy {
		return "minecraft:mossy_cobblestone", nil
	}
	return "minecraft:cobblestone", nil
}
