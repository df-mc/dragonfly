package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct {
	solid
}

// plantSelection are the plants that are picked from when a bone meal is attempted.
// TODO: Base plant selection on current biome.
var plantSelection = []world.Block{
	Flower{Type: OxeyeDaisy()},
	Flower{Type: PinkTulip()},
	Flower{Type: Cornflower()},
	Flower{Type: WhiteTulip()},
	Flower{Type: RedTulip()},
	Flower{Type: OrangeTulip()},
	Flower{Type: Dandelion()},
	Flower{Type: Poppy()},
}

// init adds extra variants of TallGrass to the plant selection.
func init() {
	for i := 0; i < 8; i++ {
		plantSelection = append(plantSelection, Fern{})
	}
	for i := 0; i < 12; i++ {
		plantSelection = append(plantSelection, ShortGrass{})
	}
}

// SoilFor ...
func (g Grass) SoilFor(block world.Block) bool {
	switch block.(type) {
	case ShortGrass, Fern, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, PinkPetals, SugarCane, DeadBush, BambooSapling, Bamboo:
		return true
	}
	return false
}

// RandomTick handles the ticking of grass, which may or may not result in the spreading of grass onto dirt.
func (g Grass) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	spreadDirt(g, pos, tx, r, -3)
}

// BoneMeal ...
func (g Grass) BoneMeal(pos cube.Pos, tx *world.Tx) (result item.BoneMealResult) {
	result = item.BoneMealResultNone
	for range 14 {
		c := pos.Add(cube.Pos{rand.IntN(6) - 3, 0, rand.IntN(6) - 3})
		above := c.Side(cube.FaceUp)
		_, air := tx.Block(above).(Air)
		_, grass := tx.Block(c).(Grass)
		if air && grass {
			tx.SetBlock(above, plantSelection[rand.IntN(len(plantSelection))], nil)
			result = item.BoneMealResultArea
		}
	}
	return
}

// BreakInfo ...
func (g Grass) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, g))
}

// CompostChance ...
func (Grass) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (Grass) EncodeItem() (name string, meta int16) {
	return "minecraft:grass_block", 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (string, map[string]any) {
	return "minecraft:grass_block", nil
}

// Till ...
func (g Grass) Till() (world.Block, bool) {
	return Farmland{}, true
}

// Shovel ...
func (g Grass) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}
