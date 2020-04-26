package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"

// Dirt is a block found abundantly in most biomes under a layer of grass blocks at the top of the normal
// world.
type Dirt struct {
	// Coarse specifies if the dirt should be off the coarse dirt variant. Grass blocks won't spread on
	// the block if set to true.
	Coarse bool
}

// BreakInfo ...
func (d Dirt) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.5,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(d, 1)),
	}
}

// EncodeItem ...
func (d Dirt) EncodeItem() (id int32, meta int16) {
	if d.Coarse {
		meta = 1
	}
	return 3, meta
}

// EncodeBlock ...
func (d Dirt) EncodeBlock() (name string, properties map[string]interface{}) {
	if d.Coarse {
		return "minecraft:dirt", map[string]interface{}{"dirt_type": "coarse"}
	}
	return "minecraft:dirt", map[string]interface{}{"dirt_type": "normal"}
}
