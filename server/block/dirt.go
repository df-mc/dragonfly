package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Dirt is a block found abundantly in most biomes under a layer of grass blocks at the top of the normal
// world.
type Dirt struct {
	solid

	// Coarse specifies if the dirt should be off the coarse dirt variant. Grass blocks won't spread on
	// the block if set to true.
	Coarse bool
}

// SoilFor ...
func (d Dirt) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, DeadBush:
		return true
	}
	return false
}

// BreakInfo ...
func (d Dirt) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(d))
}

// Till ...
func (d Dirt) Till() (world.Block, bool) {
	if d.Coarse {
		return Dirt{Coarse: false}, true
	}
	return Farmland{}, true
}

// Shovel ...
func (d Dirt) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}

// EncodeItem ...
func (d Dirt) EncodeItem() (name string, meta int16) {
	if d.Coarse {
		meta = 1
	}
	return "minecraft:dirt", meta
}

// EncodeBlock ...
func (d Dirt) EncodeBlock() (string, map[string]any) {
	if d.Coarse {
		return "minecraft:dirt", map[string]any{"dirt_type": "coarse"}
	}
	return "minecraft:dirt", map[string]any{"dirt_type": "normal"}
}
