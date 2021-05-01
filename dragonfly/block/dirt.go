package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Dirt is a block found abundantly in most biomes under a layer of grass blocks at the top of the normal
// world.
type Dirt struct {
	solid

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

// Till ...
func (d Dirt) Till() (world.Block, bool) {
	if d.Coarse {
		return Dirt{Coarse: false}, true
	}
	return Farmland{}, true
}

// EncodeItem ...
func (d Dirt) EncodeItem() (id int32, name string, meta int16) {
	if d.Coarse {
		meta = 1
	}
	return 3, "minecraft:dirt", meta
}

// EncodeBlock ...
func (d Dirt) EncodeBlock() (string, map[string]interface{}) {
	if d.Coarse {
		return "minecraft:dirt", map[string]interface{}{"dirt_type": "coarse"}
	}
	return "minecraft:dirt", map[string]interface{}{"dirt_type": "normal"}
}
