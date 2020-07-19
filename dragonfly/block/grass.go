package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct {
	noNBT
	// Path specifies if the grass was made into a path or not. If true, the block will have only 15/16th of
	// the height of a full block.
	Path bool
}

// NeighbourUpdateTick handles the turning from grass path into dirt if a block is placed on top of it.
func (g Grass) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if !g.Path {
		return
	}
	if _, air := w.Block(pos.Add(world.BlockPos{0, 1})).(Air); !air {
		// Technically vanilla doesn't always turn grass paths into dirt when a block is placed above it,
		// for example torches, but the logic doesn't make sense.
		w.SetBlock(pos, Dirt{})
	}
}

// RandomTick handles the ticking of grass, which may or may not result in the spreading of grass onto dirt.
func (g Grass) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if g.Path {
		// Don't spread grass paths.
		return
	}
	aboveLight := w.Light(pos.Add(world.BlockPos{0, 1}))
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
		spreadPos := pos.Add(world.BlockPos{r.Intn(3) - 1, r.Intn(5) - 3, r.Intn(3) - 1})
		b := w.Block(spreadPos)
		if dirt, ok := b.(Dirt); !ok || dirt.Coarse {
			continue
		}
		// Don't spread grass to places where dirt is exposed to hardly any light.
		if w.Light(spreadPos.Add(world.BlockPos{0, 1})) < 4 {
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
func (g Grass) EncodeItem() (id int32, meta int16) {
	if g.Path {
		return 198, 0
	}
	return 2, 0
}

// EncodeBlock ...
func (g Grass) EncodeBlock() (name string, properties map[string]interface{}) {
	if g.Path {
		return "minecraft:grass_path", nil
	}
	return "minecraft:grass", nil
}

// Hash ...
func (g Grass) Hash() uint64 {
	return hashGrass | (uint64(boolByte(g.Path)) << 32)
}
