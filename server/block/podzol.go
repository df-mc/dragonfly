package block

import "github.com/df-mc/dragonfly/server/world"

// Podzol is a dirt-type block that naturally blankets the surface of the giant tree taiga and bamboo jungles, along
// with their respective variants.
type Podzol struct {
	solid
}

func (p Podzol) SoilFor(block world.Block) bool {
	switch block.(type) {
	case ShortGrass, Fern, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, DeadBush, SugarCane:
		return true
	}
	return false
}

func (Podzol) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}

func (p Podzol) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, p))
}

func (Podzol) EncodeItem() (name string, meta int16) {
	return "minecraft:podzol", 0
}

func (Podzol) EncodeBlock() (string, map[string]any) {
	return "minecraft:podzol", nil
}
