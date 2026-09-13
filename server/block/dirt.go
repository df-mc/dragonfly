package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
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
	case Flower, DoubleFlower, NetherSprouts, PinkPetals, SugarCane, BambooSapling, Bamboo:
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

// spreadDirt turns the block passed into dirt in low light and otherwise spreads it onto nearby dirt. It is
// the random tick behaviour shared by grass blocks and mycelium. minY is the lowest vertical offset a spread
// attempt may target.
func spreadDirt(b world.Block, pos cube.Pos, tx *world.Tx, r *rand.Rand, minY int) {
	aboveLight := tx.Light(pos.Side(cube.FaceUp))
	if aboveLight < 4 {
		tx.SetBlock(pos, Dirt{}, nil)
		return
	}
	if aboveLight < 9 {
		return
	}

	// A single uint32 is enough as only 28 bits are needed, 7 per iteration.
	n := r.Uint32()

	for range 4 {
		x, y, z := int(n)%3, int(n>>2)%5, int(n>>5)%3
		n >>= 7

		spreadPos := pos.Add(cube.Pos{x - 1, y + minY, z - 1})
		if tx.Light(spreadPos.Side(cube.FaceUp)) < 4 {
			continue
		}
		if dirt, ok := tx.Block(spreadPos).(Dirt); !ok || dirt.Coarse {
			continue
		}
		tx.SetBlock(spreadPos, b, nil)
	}
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
