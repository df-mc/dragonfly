package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand/v2"
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
	case ShortGrass, Fern, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, PinkPetals, SugarCane, DeadBush:
		return true
	}
	return false
}

// RandomTick handles the ticking of grass, which may or may not result in the spreading of grass onto dirt.
func (g Grass) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	aboveLight := tx.Light(pos.Side(cube.FaceUp))
	if aboveLight < 4 {
		// The light above the block is too low: The grass turns to dirt.
		tx.SetBlock(pos, Dirt{}, nil)
		return
	}
	if aboveLight < 9 {
		// Don't attempt to spread if the light level is lower than 9.
		return
	}

	// Generate a single uint32 as we only need 28 bits (7 bits each iteration).
	n := r.Uint32()

	// Four attempts to spread to another block.
	for i := 0; i < 4; i++ {
		x, y, z := int(n)%3, int(n>>2)%5, int(n>>5)%3
		n >>= 7

		spreadPos := pos.Add(cube.Pos{x - 1, y - 3, z - 1})
		// Don't spread grass to locations where dirt is exposed to hardly any light.
		if tx.Light(spreadPos.Side(cube.FaceUp)) < 4 {
			continue
		}
		b := tx.Block(spreadPos)
		if dirt, ok := b.(Dirt); !ok || dirt.Coarse {
			continue
		}
		tx.SetBlock(spreadPos, g, nil)
	}
}

// BoneMeal ...
func (g Grass) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	for i := 0; i < 14; i++ {
		c := pos.Add(cube.Pos{rand.IntN(6) - 3, 0, rand.IntN(6) - 3})
		above := c.Side(cube.FaceUp)
		_, air := tx.Block(above).(Air)
		_, grass := tx.Block(c).(Grass)
		if air && grass {
			tx.SetBlock(above, plantSelection[rand.IntN(len(plantSelection))], nil)
		}
	}

	return false
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
