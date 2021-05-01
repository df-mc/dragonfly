package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct {
	solid
}

// RandomTick handles the ticking of grass, which may or may not result in the spreading of grass onto dirt.
func (g Grass) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	aboveLight := w.Light(pos.Add(cube.Pos{0, 1}))
	if aboveLight < 4 {
		// The light above the block is too low: The grass turns to dirt.
		w.SetBlock(pos, Dirt{})
		return
	}
	if aboveLight < 9 {
		// Don't attempt to spread if the light level is lower than 9.
		return
	}
	// Four attempts to spread to another block.
	for i := 0; i < 4; i++ {
		spreadPos := pos.Add(cube.Pos{r.Intn(3) - 1, r.Intn(5) - 3, r.Intn(3) - 1})
		b := w.Block(spreadPos)
		if dirt, ok := b.(Dirt); !ok || dirt.Coarse {
			continue
		}
		// Don't spread grass to places where dirt is exposed to hardly any light.
		if w.Light(spreadPos.Add(cube.Pos{0, 1})) < 4 {
			continue
		}
		w.SetBlock(spreadPos, g)
	}
}

// BreakInfo ...
func (g Grass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(Dirt{}, 1)),
	}
}

// EncodeItem ...
func (Grass) EncodeItem() (id int32, name string, meta int16) {
	return 2, "minecraft:grass", 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:grass", nil
}

// Till ...
func (g Grass) Till() (world.Block, bool) {
	return Farmland{}, true
}

// Shovel ...
func (g Grass) Shovel() (world.Block, bool) {
	return GrassPath{}, true
}
