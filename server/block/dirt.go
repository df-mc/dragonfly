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
	case ShortGrass, Fern, DoubleTallGrass, DeadBush:
		return !d.Coarse
	case Flower, DoubleFlower, NetherSprouts, SugarCane:
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
		return "minecraft:coarse_dirt", 0
	}
	return "minecraft:dirt", 0
}

// EncodeBlock ...
func (d Dirt) EncodeBlock() (string, map[string]any) {
	if d.Coarse {
		return "minecraft:coarse_dirt", nil
	}
	return "minecraft:dirt", nil
}

// supportsVegetation checks if the vegetation can exist on the block.
func supportsVegetation(vegetation, block world.Block) bool {
	soil, ok := block.(Soil)
	return ok && soil.SoilFor(vegetation)
}

// Soil represents a block that can support vegetation.
type Soil interface {
	// SoilFor returns whether the vegetation can exist on the block.
	SoilFor(world.Block) bool
}
