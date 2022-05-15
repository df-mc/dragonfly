package block

import "github.com/df-mc/dragonfly/server/world"

// Podzol is a dirt-type block that naturally blankets the surface of the giant tree taiga and bamboo jungles, along
// with their respective variants.
type Podzol struct {
	solid
}

// SoilFor ...
func (p Podzol) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, DeadBush:
		return true
	}
	return false
}

// Shovel ...
func (Podzol) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}

// BreakInfo ...
func (p Podzol) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, p))
}

// EncodeItem ...
func (Podzol) EncodeItem() (name string, meta int16) {
	return "minecraft:podzol", 0
}

// EncodeBlock ...
func (Podzol) EncodeBlock() (string, map[string]any) {
	return "minecraft:podzol", nil
}
